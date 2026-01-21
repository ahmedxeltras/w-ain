$args = @("*.ts", "*.tsx", "*.js", "*.cjs", "*.html", "*.md", "*.mdx", "*.json")
$files = Get-ChildItem -Recurse -Include $args -Exclude "package-lock.json"
foreach ($file in $files) {
	# Skip node_modules and .git
	if ($file.FullName -like "*\node_modules\*") { continue }
	if ($file.FullName -like "*\.git\*") { continue }
	if ($file.FullName -like "*\.sso\*") { continue }

	try {
		$content = Get-Content $file.FullName -Raw -ErrorAction Stop
		$newContent = $content
        
		# Display Texts
		$newContent = $newContent -replace "Wave Terminal", "Ain Term"
		$newContent = $newContent -replace "WaveApp", "AinApp"
		# Be careful with just "Wave". "Wave" is generic. But usually "Wave" in this repo means the app.
		# "Wave" -> "Ain Term" might be too aggressive? "Wave" -> "Ain"?
		# "WaveInitOpts" -> "AinTermInitOpts"?
		# I will replace "WaveInitOpts" -> "AinInitOpts"
		$newContent = $newContent -replace "WaveInitOpts", "AinInitOpts"
		$newContent = $newContent -replace "reinitWave", "reinitAin"
		$newContent = $newContent -replace "initWave", "initAin"
        
		# Config / Env
		$newContent = $newContent -replace "WAVETERM_", "AINTERM_"
		$newContent = $newContent -replace "dev.commandline.waveterm", "dev.commandline.ainterm"

		# URLs/Paths
		$newContent = $newContent -replace "\.waveterm", ".ainterm"
        
		# Specific component/file renames in imports
		# frontend/wave.ts -> frontend/ainterm.ts
		$newContent = $newContent -replace "frontend/wave\.ts", "frontend/ainterm.ts"
		$newContent = $newContent -replace "frontend/wave", "frontend/ainterm"
        
		# waveconfig -> ainconfig
		$newContent = $newContent -replace "waveconfig", "ainconfig"

		# waveai -> ainai (including waveaivisual -> ainaivisual)
		$newContent = $newContent -replace "waveai", "ainai"
		$newContent = $newContent -replace "WaveAi", "AinAi"
		$newContent = $newContent -replace "WaveAI", "AinAI"

		# waveutil -> ainutil
		# $newContent = $newContent -replace "waveutil", "ainutil"
		# Wait, if I use "@/util/waveutil" it might be safer
		$newContent = $newContent -replace "/waveutil", "/ainutil"

		# wave-ready -> ain-ready
		$newContent = $newContent -replace "wave-ready", "ain-ready"

		# Title
		if ($file.Extension -eq ".html") {
			$newContent = $newContent -replace "<title>Wave</title>", "<title>Ain Term</title>"
		}

		if ($content -ne $newContent) {
			Set-Content -Path $file.FullName -Value $newContent -NoNewline
			Write-Host "Updated $($file.Name)"
		}
	}
 catch {
		Write-Host "Error processing $($file.Name): $_"
	}
}
