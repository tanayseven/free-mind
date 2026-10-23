$ErrorActionPreference = 'Stop'

# Remove the daemon the app installed on first launch.
$service = Get-Service FreeMindDaemon -ErrorAction SilentlyContinue
if ($service) {
  Stop-Service FreeMindDaemon -Force
  sc.exe delete FreeMindDaemon | Out-Null
}
Remove-Item 'C:\Program Files\Free Mind\free-mind-daemon.exe' -Force -ErrorAction SilentlyContinue

# This script replaces Chocolatey's automatic uninstaller, so remove the MSI.
[array]$keys = Get-UninstallRegistryKey -SoftwareName 'Free Mind'
foreach ($key in $keys) {
  Uninstall-ChocolateyPackage -PackageName $env:ChocolateyPackageName -FileType 'msi' `
    -SilentArgs "$($key.PSChildName) /qn /norestart" -ValidExitCodes @(0, 1605, 1614, 1641, 3010)
}
