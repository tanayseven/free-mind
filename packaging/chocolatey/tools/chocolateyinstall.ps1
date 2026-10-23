$ErrorActionPreference = 'Stop'

$packageArgs = @{
  packageName    = $env:ChocolateyPackageName
  fileType       = 'msi'
  url64bit       = 'https://github.com/tanayseven/free-mind/releases/download/v{{VERSION}}/Free.Mind_{{VERSION}}_x64_en-US.msi'
  checksum64     = '{{SHA256_MSI_X64}}'
  checksumType64 = 'sha256'
  silentArgs     = '/qn /norestart'
  validExitCodes = @(0, 1641, 3010)
}
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
  $packageArgs.url64bit = 'https://github.com/tanayseven/free-mind/releases/download/v{{VERSION}}/Free.Mind_{{VERSION}}_arm64_en-US.msi'
  $packageArgs.checksum64 = '{{SHA256_MSI_ARM64}}'
}

Install-ChocolateyPackage @packageArgs
