// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"runtime"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/wavetermdev/ainterm/pkg/ainbase"
	"github.com/wavetermdev/ainterm/pkg/aincloud"
	"github.com/wavetermdev/ainterm/pkg/ainconfig"
	"github.com/wavetermdev/ainterm/pkg/aincore"
	"github.com/wavetermdev/ainterm/pkg/ainobj"
	"github.com/wavetermdev/ainterm/pkg/ainps"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshclient"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshremote"
	"github.com/wavetermdev/ainterm/pkg/ainshrpc/wshserver"
	"github.com/wavetermdev/ainterm/pkg/ainshutil"
	"github.com/wavetermdev/ainterm/pkg/ainstore"
	"github.com/wavetermdev/ainterm/pkg/aiusechat"
	"github.com/wavetermdev/ainterm/pkg/authkey"
	"github.com/wavetermdev/ainterm/pkg/blockcontroller"
	"github.com/wavetermdev/ainterm/pkg/blocklogger"
	"github.com/wavetermdev/ainterm/pkg/filebackup"
	"github.com/wavetermdev/ainterm/pkg/filestore"
	"github.com/wavetermdev/ainterm/pkg/panichandler"
	"github.com/wavetermdev/ainterm/pkg/remote/conncontroller"
	"github.com/wavetermdev/ainterm/pkg/remote/fileshare/wshfs"
	"github.com/wavetermdev/ainterm/pkg/secretstore"
	"github.com/wavetermdev/ainterm/pkg/service"
	"github.com/wavetermdev/ainterm/pkg/telemetry"
	"github.com/wavetermdev/ainterm/pkg/telemetry/telemetrydata"
	"github.com/wavetermdev/ainterm/pkg/util/shellutil"
	"github.com/wavetermdev/ainterm/pkg/util/sigutil"
	"github.com/wavetermdev/ainterm/pkg/util/utilfn"
	"github.com/wavetermdev/ainterm/pkg/web"
	"github.com/wavetermdev/ainterm/pkg/wslconn"

	"net/http"
	_ "net/http/pprof"
)

// these are set at build time
var WaveVersion = "0.0.0"
var BuildTime = "0"

const InitialTelemetryWait = 10 * time.Second
const TelemetryTick = 2 * time.Minute
const TelemetryInterval = 4 * time.Hour
const TelemetryInitialCountsWait = 5 * time.Second
const TelemetryCountsInterval = 1 * time.Hour
const BackupCleanupTick = 2 * time.Minute
const BackupCleanupInterval = 4 * time.Hour
const InitialDiagnosticWait = 5 * time.Minute
const DiagnosticTick = 10 * time.Minute

var shutdownOnce sync.Once

func init() {
	envFilePath := os.Getenv("AINTERM_ENVFILE")
	if envFilePath != "" {
		log.Printf("applying env file: %s\n", envFilePath)
		_ = godotenv.Load(envFilePath)
	}
}

func doShutdown(reason string) {
	shutdownOnce.Do(func() {
		log.Printf("shutting down: %s\n", reason)
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		go blockcontroller.StopAllBlockControllers()
		shutdownActivityUpdate()
		sendTelemetryWrapper()
		// TODO deal with flush in progress
		clearTempFiles()
		filestore.WFS.FlushCache(ctx)
		watcher := ainconfig.GetWatcher()
		if watcher != nil {
			watcher.Close()
		}
		time.Sleep(500 * time.Millisecond)
		log.Printf("shutdown complete\n")
		os.Exit(0)
	})
}

// watch stdin, kill server if stdin is closed
func stdinReadWatch() {
	defer func() {
		panichandler.PanicHandler("stdinReadWatch", recover())
	}()
	buf := make([]byte, 1024)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			doShutdown(fmt.Sprintf("stdin closed/error (%v)", err))
			break
		}
	}
}

func startConfigWatcher() {
	watcher := ainconfig.GetWatcher()
	if watcher != nil {
		watcher.Start()
	}
}

func telemetryLoop() {
	defer func() {
		panichandler.PanicHandler("telemetryLoop", recover())
	}()
	var nextSend int64
	time.Sleep(InitialTelemetryWait)
	for {
		if time.Now().Unix() > nextSend {
			nextSend = time.Now().Add(TelemetryInterval).Unix()
			sendTelemetryWrapper()
		}
		time.Sleep(TelemetryTick)
	}
}

func diagnosticLoop() {
	defer func() {
		panichandler.PanicHandler("diagnosticLoop", recover())
	}()
	if os.Getenv("AINTERM_NOPING") != "" {
		log.Printf("AINTERM_NOPING set, disabling diagnostic ping\n")
		return
	}
	var lastSentDate string
	time.Sleep(InitialDiagnosticWait)
	for {
		currentDate := time.Now().Format("2006-01-02")
		if lastSentDate == "" || lastSentDate != currentDate {
			if sendDiagnosticPing() {
				lastSentDate = currentDate
			}
		}
		time.Sleep(DiagnosticTick)
	}
}

func sendDiagnosticPing() bool {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	rpcClient := wshclient.GetBareRpcClient()
	isOnline, err := wshclient.NetworkOnlineCommand(rpcClient, &ainshrpc.RpcOpts{Route: "electron", Timeout: 2000})
	if err != nil || !isOnline {
		return false
	}
	clientData, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		return false
	}
	if clientData == nil {
		return false
	}
	usageTelemetry := telemetry.IsTelemetryEnabled()
	aincloud.SendDiagnosticPing(ctx, clientData.OID, usageTelemetry)
	return true
}

func setupTelemetryConfigHandler() {
	watcher := ainconfig.GetWatcher()
	if watcher == nil {
		return
	}
	currentConfig := watcher.GetFullConfig()
	currentTelemetryEnabled := currentConfig.Settings.TelemetryEnabled

	watcher.RegisterUpdateHandler(func(newConfig ainconfig.FullConfigType) {
		newTelemetryEnabled := newConfig.Settings.TelemetryEnabled
		if newTelemetryEnabled != currentTelemetryEnabled {
			currentTelemetryEnabled = newTelemetryEnabled
			aincore.GoSendNoTelemetryUpdate(newTelemetryEnabled)
		}
	})
}

func backupCleanupLoop() {
	defer func() {
		panichandler.PanicHandler("backupCleanupLoop", recover())
	}()
	var nextCleanup int64
	for {
		if time.Now().Unix() > nextCleanup {
			nextCleanup = time.Now().Add(BackupCleanupInterval).Unix()
			err := filebackup.CleanupOldBackups()
			if err != nil {
				log.Printf("error cleaning up old backups: %v\n", err)
			}
		}
		time.Sleep(BackupCleanupTick)
	}
}

func panicTelemetryHandler(panicName string) {
	activity := ainshrpc.ActivityUpdate{NumPanics: 1}
	err := telemetry.UpdateActivity(context.Background(), activity)
	if err != nil {
		log.Printf("error updating activity (panicTelemetryHandler): %v\n", err)
	}
	telemetry.RecordTEvent(context.Background(), telemetrydata.MakeTEvent("debug:panic", telemetrydata.TEventProps{
		PanicType: panicName,
	}))
}

func sendTelemetryWrapper() {
	defer func() {
		panichandler.PanicHandler("sendTelemetryWrapper", recover())
	}()
	ctx, cancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFn()
	beforeSendActivityUpdate(ctx)
	client, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		log.Printf("[error] getting client data for telemetry: %v\n", err)
		return
	}
	err = aincloud.SendAllTelemetry(client.OID)
	if err != nil {
		log.Printf("[error] sending telemetry: %v\n", err)
	}
}

func updateTelemetryCounts(lastCounts telemetrydata.TEventProps) telemetrydata.TEventProps {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	var props telemetrydata.TEventProps
	props.CountBlocks, _ = ainstore.DBGetCount[*ainobj.Block](ctx)
	props.CountTabs, _ = ainstore.DBGetCount[*ainobj.Tab](ctx)
	props.CountWindows, _ = ainstore.DBGetCount[*ainobj.Window](ctx)
	props.CountWorkspaces, _, _ = ainstore.DBGetWSCounts(ctx)
	props.CountSSHConn = conncontroller.GetNumSSHHasConnected()
	props.CountWSLConn = wslconn.GetNumWSLHasConnected()
	props.CountViews, _ = ainstore.DBGetBlockViewCounts(ctx)

	fullConfig := ainconfig.GetWatcher().GetFullConfig()
	customWidgets := fullConfig.CountCustomWidgets()
	customAIPresets := fullConfig.CountCustomAIPresets()
	customSettings := ainconfig.CountCustomSettings()
	customAIModes := fullConfig.CountCustomAIModes()

	props.UserSet = &telemetrydata.TEventUserProps{
		SettingsCustomWidgets:   customWidgets,
		SettingsCustomAIPresets: customAIPresets,
		SettingsCustomSettings:  customSettings,
		SettingsCustomAIModes:   customAIModes,
	}

	secretsCount, err := secretstore.CountSecrets()
	if err == nil {
		props.UserSet.SettingsSecretsCount = secretsCount
	}

	if utilfn.CompareAsMarshaledJson(props, lastCounts) {
		return lastCounts
	}
	tevent := telemetrydata.MakeTEvent("app:counts", props)
	err = telemetry.RecordTEvent(ctx, tevent)
	if err != nil {
		log.Printf("error recording counts tevent: %v\n", err)
	}
	return props
}

func updateTelemetryCountsLoop() {
	defer func() {
		panichandler.PanicHandler("updateTelemetryCountsLoop", recover())
	}()
	var nextSend int64
	var lastCounts telemetrydata.TEventProps
	time.Sleep(TelemetryInitialCountsWait)
	for {
		if time.Now().Unix() > nextSend {
			nextSend = time.Now().Add(TelemetryCountsInterval).Unix()
			lastCounts = updateTelemetryCounts(lastCounts)
		}
		time.Sleep(TelemetryTick)
	}
}

func beforeSendActivityUpdate(ctx context.Context) {
	activity := ainshrpc.ActivityUpdate{}
	activity.NumTabs, _ = ainstore.DBGetCount[*ainobj.Tab](ctx)
	activity.NumBlocks, _ = ainstore.DBGetCount[*ainobj.Block](ctx)
	activity.Blocks, _ = ainstore.DBGetBlockViewCounts(ctx)
	activity.NumWindows, _ = ainstore.DBGetCount[*ainobj.Window](ctx)
	activity.NumSSHConn = conncontroller.GetNumSSHHasConnected()
	activity.NumWSLConn = wslconn.GetNumWSLHasConnected()
	activity.NumWSNamed, activity.NumWS, _ = ainstore.DBGetWSCounts(ctx)
	err := telemetry.UpdateActivity(ctx, activity)
	if err != nil {
		log.Printf("error updating before activity: %v\n", err)
	}
}

func startupActivityUpdate(firstLaunch bool) {
	defer func() {
		panichandler.PanicHandler("startupActivityUpdate", recover())
	}()
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()
	activity := ainshrpc.ActivityUpdate{Startup: 1}
	err := telemetry.UpdateActivity(ctx, activity) // set at least one record into activity (don't use go routine wrap here)
	if err != nil {
		log.Printf("error updating startup activity: %v\n", err)
	}
	autoUpdateChannel := telemetry.AutoUpdateChannel()
	autoUpdateEnabled := telemetry.IsAutoUpdateEnabled()
	shellType, shellVersion, shellErr := shellutil.DetectShellTypeAndVersion()
	if shellErr != nil {
		shellType = "error"
		shellVersion = ""
	}
	userSetOnce := &telemetrydata.TEventUserProps{
		ClientInitialVersion: "v" + WaveVersion,
	}
	tosTs := telemetry.GetTosAgreedTs()
	var cohortTime time.Time
	if tosTs > 0 {
		cohortTime = time.UnixMilli(tosTs)
	} else {
		cohortTime = time.Now()
	}
	cohortMonth := cohortTime.Format("2006-01")
	year, week := cohortTime.ISOWeek()
	cohortISOWeek := fmt.Sprintf("%04d-W%02d", year, week)
	userSetOnce.CohortMonth = cohortMonth
	userSetOnce.CohortISOWeek = cohortISOWeek
	fullConfig := ainconfig.GetWatcher().GetFullConfig()
	props := telemetrydata.TEventProps{
		UserSet: &telemetrydata.TEventUserProps{
			ClientVersion:       "v" + ainbase.WaveVersion,
			ClientBuildTime:     ainbase.BuildTime,
			ClientArch:          ainbase.ClientArch(),
			ClientOSRelease:     ainbase.UnameKernelRelease(),
			ClientIsDev:         ainbase.IsDevMode(),
			AutoUpdateChannel:   autoUpdateChannel,
			AutoUpdateEnabled:   autoUpdateEnabled,
			LocalShellType:      shellType,
			LocalShellVersion:   shellVersion,
			SettingsTransparent: fullConfig.Settings.WindowTransparent,
		},
		UserSetOnce: userSetOnce,
	}
	if firstLaunch {
		props.AppFirstLaunch = true
	}
	tevent := telemetrydata.MakeTEvent("app:startup", props)
	err = telemetry.RecordTEvent(ctx, tevent)
	if err != nil {
		log.Printf("error recording startup event: %v\n", err)
	}
}

func shutdownActivityUpdate() {
	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFn()
	activity := ainshrpc.ActivityUpdate{Shutdown: 1}
	err := telemetry.UpdateActivity(ctx, activity) // do NOT use the go routine wrap here (this needs to be synchronous)
	if err != nil {
		log.Printf("error updating shutdown activity: %v\n", err)
	}
	err = telemetry.TruncateActivityTEventForShutdown(ctx)
	if err != nil {
		log.Printf("error truncating activity t-event for shutdown: %v\n", err)
	}
	tevent := telemetrydata.MakeTEvent("app:shutdown", telemetrydata.TEventProps{})
	err = telemetry.RecordTEvent(ctx, tevent)
	if err != nil {
		log.Printf("error recording shutdown event: %v\n", err)
	}
}

func createMainWshClient() {
	rpc := wshserver.GetMainRpcClient()
	wshfs.RpcClient = rpc
	ainshutil.DefaultRouter.RegisterTrustedLeaf(rpc, ainshutil.DefaultRoute)
	ainps.Broker.SetClient(ainshutil.DefaultRouter)
	localConnWsh := ainshutil.MakeWshRpc(ainshrpc.RpcContext{Conn: ainshrpc.LocalConnName}, &wshremote.ServerImpl{}, "conn:local")
	go wshremote.RunSysInfoLoop(localConnWsh, ainshrpc.LocalConnName)
	ainshutil.DefaultRouter.RegisterTrustedLeaf(localConnWsh, ainshutil.MakeConnectionRouteId(ainshrpc.LocalConnName))
}

func grabAndRemoveEnvVars() error {
	err := authkey.SetAuthKeyFromEnv()
	if err != nil {
		return fmt.Errorf("setting auth key: %v", err)
	}
	err = ainbase.CacheAndRemoveEnvVars()
	if err != nil {
		return err
	}
	err = aincloud.CacheAndRemoveEnvVars()
	if err != nil {
		return err
	}

	// Remove WAVETERM env vars that leak from prod => dev
	os.Unsetenv("AINTERM_CLIENTID")
	os.Unsetenv("AINTERM_WORKSPACEID")
	os.Unsetenv("AINTERM_TABID")
	os.Unsetenv("AINTERM_BLOCKID")
	os.Unsetenv("AINTERM_CONN")
	os.Unsetenv("AINTERM_JWT")
	os.Unsetenv("AINTERM_VERSION")

	return nil
}

func clearTempFiles() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	client, err := ainstore.DBGetSingleton[*ainobj.Client](ctx)
	if err != nil {
		return fmt.Errorf("error getting client: %v", err)
	}
	filestore.WFS.DeleteZone(ctx, client.TempOID)
	return nil
}

func maybeStartPprofServer() {
	settings := ainconfig.GetWatcher().GetFullConfig().Settings
	if settings.DebugPprofMemProfileRate != nil {
		runtime.MemProfileRate = *settings.DebugPprofMemProfileRate
		log.Printf("set runtime.MemProfileRate to %d\n", runtime.MemProfileRate)
	}
	if settings.DebugPprofPort == nil {
		return
	}
	pprofPort := *settings.DebugPprofPort
	if pprofPort < 1 || pprofPort > 65535 {
		log.Printf("[error] debug:pprofport must be between 1 and 65535, got %d\n", pprofPort)
		return
	}
	go func() {
		addr := fmt.Sprintf("localhost:%d", pprofPort)
		log.Printf("starting pprof server on %s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("[error] pprof server failed: %v\n", err)
		}
	}()
}

func main() {
	log.SetFlags(0) // disable timestamp since electron's winston logger already wraps with timestamp
	log.SetPrefix("[wavesrv] ")
	ainbase.WaveVersion = WaveVersion
	ainbase.BuildTime = BuildTime
	ainshutil.DefaultRouter = ainshutil.NewWshRouter()
	ainshutil.DefaultRouter.SetAsRootRouter()

	err := grabAndRemoveEnvVars()
	if err != nil {
		log.Printf("[error] %v\n", err)
		return
	}
	err = service.ValidateServiceMap()
	if err != nil {
		log.Printf("error validating service map: %v\n", err)
		return
	}
	err = ainbase.EnsureWaveDataDir()
	if err != nil {
		log.Printf("error ensuring wave home dir: %v\n", err)
		return
	}
	err = ainbase.EnsureWaveDBDir()
	if err != nil {
		log.Printf("error ensuring wave db dir: %v\n", err)
		return
	}
	err = ainbase.EnsureWaveConfigDir()
	if err != nil {
		log.Printf("error ensuring wave config dir: %v\n", err)
		return
	}

	// TODO: rather than ensure this dir exists, we should let the editor recursively create parent dirs on save
	err = ainbase.EnsureWavePresetsDir()
	if err != nil {
		log.Printf("error ensuring wave presets dir: %v\n", err)
		return
	}
	err = ainbase.EnsureWaveCachesDir()
	if err != nil {
		log.Printf("error ensuring wave caches dir: %v\n", err)
		return
	}
	waveLock, err := ainbase.AcquireWaveLock()
	if err != nil {
		log.Printf("error acquiring wave lock (another instance of Wave is likely running): %v\n", err)
		return
	}
	defer func() {
		err = waveLock.Close()
		if err != nil {
			log.Printf("error releasing wave lock: %v\n", err)
		}
	}()
	log.Printf("wave version: %s (%s)\n", WaveVersion, BuildTime)
	log.Printf("wave data dir: %s\n", ainbase.GetWaveDataDir())
	log.Printf("wave config dir: %s\n", ainbase.GetWaveConfigDir())
	err = filestore.InitFilestore()
	if err != nil {
		log.Printf("error initializing filestore: %v\n", err)
		return
	}
	err = ainstore.InitWStore()
	if err != nil {
		log.Printf("error initializing wstore: %v\n", err)
		return
	}
	panichandler.PanicTelemetryHandler = panicTelemetryHandler
	go func() {
		defer func() {
			panichandler.PanicHandler("InitCustomShellStartupFiles", recover())
		}()
		err := shellutil.InitCustomShellStartupFiles()
		if err != nil {
			log.Printf("error initializing wsh and shell-integration files: %v\n", err)
		}
	}()
	firstLaunch, err := aincore.EnsureInitialData()
	if err != nil {
		log.Printf("error ensuring initial data: %v\n", err)
		return
	}
	if firstLaunch {
		log.Printf("first launch detected")
	}
	err = clearTempFiles()
	if err != nil {
		log.Printf("error clearing temp files: %v\n", err)
		return
	}
	err = aincore.InitMainServer()
	if err != nil {
		log.Printf("error initializing mainserver: %v\n", err)
		return
	}

	err = shellutil.FixupWaveZshHistory()
	if err != nil {
		log.Printf("error fixing up wave zsh history: %v\n", err)
	}
	createMainWshClient()
	sigutil.InstallShutdownSignalHandlers(doShutdown)
	sigutil.InstallSIGUSR1Handler()
	startConfigWatcher()
	aiusechat.InitAIModeConfigWatcher()
	maybeStartPprofServer()
	go stdinReadWatch()
	go telemetryLoop()
	go diagnosticLoop()
	setupTelemetryConfigHandler()
	go updateTelemetryCountsLoop()
	go backupCleanupLoop()
	go startupActivityUpdate(firstLaunch) // must be after startConfigWatcher()
	blocklogger.InitBlockLogger()
	go func() {
		defer func() {
			panichandler.PanicHandler("GetSystemSummary", recover())
		}()
		ainbase.GetSystemSummary()
	}()

	webListener, err := web.MakeTCPListener("web")
	if err != nil {
		log.Printf("error creating web listener: %v\n", err)
		return
	}
	wsListener, err := web.MakeTCPListener("websocket")
	if err != nil {
		log.Printf("error creating websocket listener: %v\n", err)
		return
	}
	go web.RunWebSocketServer(wsListener)
	unixListener, err := web.MakeUnixListener()
	if err != nil {
		log.Printf("error creating unix listener: %v\n", err)
		return
	}
	go func() {
		if BuildTime == "" {
			BuildTime = "0"
		}
		// use fmt instead of log here to make sure it goes directly to stderr
		fmt.Fprintf(os.Stderr, "WAVESRV-ESTART ws:%s web:%s version:%s buildtime:%s\n", wsListener.Addr(), webListener.Addr(), WaveVersion, BuildTime)
	}()
	go ainshutil.RunWshRpcOverListener(unixListener)
	web.RunWebServer(webListener) // blocking
	runtime.KeepAlive(waveLock)
}
