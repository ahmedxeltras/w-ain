$frontendDir = "c:\Users\Ahmed Eltras\Desktop\waveterm\frontend"

# Define renames
$renames = @(
	@{ Path = "app\view\waveconfig"; NewName = "ainconfig" },
	@{ Path = "app\view\waveai"; NewName = "ainai" },
	@{ Path = "app\aipanel\waveai-focus-utils.ts"; NewName = "ainai-focus-utils.ts" },
	@{ Path = "app\aipanel\waveai-model.tsx"; NewName = "ainai-model.tsx" },
	@{ Path = "util\waveutil.ts"; NewName = "ainutil.ts" }
)

foreach ($rename in $renames) {
	$fullPath = Join-Path $frontendDir $rename.Path
	if (Test-Path $fullPath) {
		$parent = Split-Path $fullPath
		$dest = Join-Path $parent $rename.NewName
		Write-Host "Renaming $fullPath to $dest"
		Move-Item -Path $fullPath -Destination $dest -Force
	}
 else {
		Write-Host "Warning: Path not found: $fullPath"
	}
}

# Rename files within directories that were just renamed
$dirRenames = @(
	@{ Dir = "app\view\ainconfig"; OldPart = "waveconfig"; NewPart = "ainconfig" },
	@{ Dir = "app\view\ainconfig"; OldPart = "waveai"; NewPart = "ainai" }, # for waveaivisual.tsx
	@{ Dir = "app\view\ainai"; OldPart = "waveai"; NewPart = "ainai" }
)

foreach ($dr in $dirRenames) {
	$fullDir = Join-Path $frontendDir $dr.Dir
	if (Test-Path $fullDir) {
		Get-ChildItem -Path $fullDir -Filter "*$($dr.OldPart)*" | ForEach-Object {
			$newName = $_.Name.Replace($dr.OldPart, $dr.NewPart)
			$dest = Join-Path $fullDir $newName
			Write-Host "Renaming $($_.FullName) to $dest"
			Move-Item -Path $_.FullName -Destination $dest -Force
		}
	}
}
