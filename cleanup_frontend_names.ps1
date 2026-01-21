$args = @("*.ts", "*.tsx")
$files = Get-ChildItem -Path "c:\Users\Ahmed Eltras\Desktop\waveterm\frontend" -Recurse -Include $args
Write-Host "Found $($files.Count) files."
foreach ($file in $files) {
	if ($file.FullName -like "*\node_modules\*") { continue }
	# Write-Host "Processing $($file.FullName)"
	$content = [System.IO.File]::ReadAllText($file.FullName)
	$newContent = $content

	# Fix class/component names
	$newContent = $newContent -creplace "ainconfigViewModel", "AinConfigViewModel"
	$newContent = $newContent -creplace "ainconfigView", "AinConfigView"
	$newContent = $newContent -creplace "ainaiModel", "AinAiModel"
	$newContent = $newContent -creplace "ainai", "AinAi"
	$newContent = $newContent -creplace "ainaiPromptMessageType", "AinAiPromptMessageType"
	$newContent = $newContent -creplace "ainaiOptsType", "AinAiOptsType"
	$newContent = $newContent -creplace "ainaiStreamRequest", "AinAiStreamRequest"
	$newContent = $newContent -creplace "StreamainaiCommand", "StreamAinAiCommand"
	$newContent = $newContent -creplace "SaveainaiData", "SaveAinAiData"

	# Fix display strings
	$newContent = $newContent -replace "Wave AI", "Ain AI"
	$newContent = $newContent -replace "Wave Config", "Ain Config"
	$newContent = $newContent -replace "Wave's AI Proxy", "Ain AI Proxy"
    
	if ($content -ne $newContent) {
		[System.IO.File]::WriteAllText($file.FullName, $newContent)
		Write-Host "Updated $($file.FullName)"
	}
}
