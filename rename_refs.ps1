$files = Get-ChildItem -Recurse -Filter "*.go"
foreach ($file in $files) {
	try {
		$content = Get-Content $file.FullName -Raw -ErrorAction Stop
		$newContent = $content
        
		# Package qualifiers (must come before imports)
		$newContent = $newContent -replace "wavebase\.", "ainbase."
		$newContent = $newContent -replace "waveai\.", "ainai."
		$newContent = $newContent -replace "waveapp\.", "ainapp."
		$newContent = $newContent -replace "waveappstore\.", "ainappstore."
		$newContent = $newContent -replace "waveapputil\.", "ainapputil."
		$newContent = $newContent -replace "wavejwt\.", "ainjwt."
		$newContent = $newContent -replace "waveobj\.", "ainobj."
		$newContent = $newContent -replace "wcloud\.", "aincloud."
		$newContent = $newContent -replace "wconfig\.", "ainconfig."
		$newContent = $newContent -replace "wcore\.", "aincore."
		$newContent = $newContent -replace "wps\.", "ainps."
		$newContent = $newContent -replace "wshrpc\.", "ainshrpc."
		$newContent = $newContent -replace "wshutil\.", "ainshutil."
		$newContent = $newContent -replace "wstore\.", "ainstore."
        
		# Module path
		$newContent = $newContent -replace "github.com/wavetermdev/waveterm", "github.com/wavetermdev/ainterm"
        
		# Imports & Paths
		$newContent = $newContent -replace "/pkg/waveai", "/pkg/ainai"
		$newContent = $newContent -replace "/pkg/waveapp", "/pkg/ainapp"
		$newContent = $newContent -replace "/pkg/waveappstore", "/pkg/ainappstore"
		$newContent = $newContent -replace "/pkg/waveapputil", "/pkg/ainapputil"
		$newContent = $newContent -replace "/pkg/wavebase", "/pkg/ainbase"
		$newContent = $newContent -replace "/pkg/wavejwt", "/pkg/ainjwt"
		$newContent = $newContent -replace "/pkg/waveobj", "/pkg/ainobj"
		$newContent = $newContent -replace "/pkg/wcloud", "/pkg/aincloud"
		$newContent = $newContent -replace "/pkg/wconfig", "/pkg/ainconfig"
		$newContent = $newContent -replace "/pkg/wcore", "/pkg/aincore"
		$newContent = $newContent -replace "/pkg/wps", "/pkg/ainps"
		$newContent = $newContent -replace "/pkg/wshrpc", "/pkg/ainshrpc"
		$newContent = $newContent -replace "/pkg/wshutil", "/pkg/ainshutil"
		$newContent = $newContent -replace "/pkg/wstore", "/pkg/ainstore"
        
		# cmd/wsh -> cmd/ainsh import
		$newContent = $newContent -replace "/cmd/wsh", "/cmd/ainsh"

		# Package declarations
		$newContent = $newContent -replace "package waveai", "package ainai"
		$newContent = $newContent -replace "package waveapp", "package ainapp"
		$newContent = $newContent -replace "package waveappstore", "package ainappstore"
		$newContent = $newContent -replace "package waveapputil", "package ainapputil"
		$newContent = $newContent -replace "package wavebase", "package ainbase"
		$newContent = $newContent -replace "package wavejwt", "package ainjwt"
		$newContent = $newContent -replace "package waveobj", "package ainobj"
		$newContent = $newContent -replace "package wcloud", "package aincloud"
		$newContent = $newContent -replace "package wconfig", "package ainconfig"
		$newContent = $newContent -replace "package wcore", "package aincore"
		$newContent = $newContent -replace "package wps", "package ainps"
		$newContent = $newContent -replace "package wshrpc", "package ainshrpc"
		$newContent = $newContent -replace "package wshutil", "package ainshutil"
		$newContent = $newContent -replace "package wstore", "package ainstore"
        
		# Environment Variables
		$newContent = $newContent -replace "WAVETERM_", "AINTERM_"
        
		# Dot Directories
		$newContent = $newContent -replace "\.waveterm", ".ainterm"

		if ($content -ne $newContent) {
			Set-Content -Path $file.FullName -Value $newContent -NoNewline
			Write-Host "Updated $($file.Name)"
		}
	}
 catch {
		Write-Host "Error processing $($file.Name): $_"
	}
}
