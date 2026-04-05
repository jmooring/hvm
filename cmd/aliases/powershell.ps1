# Hugo Version Manager: override path to the hugo executable.
function Hugo-Override {
    $hvm_show_status = $true
    $HVMBin = hvm status --printExecPathCached 2>$null
    if ($HVMBin) {
        if ($hvm_show_status) {
            $InformationPreference = 'Continue'
            Write-Host "Hugo version management is enabled in this directory."
            Write-Host "Run 'hvm status' for details, or 'hvm disable' to disable.`n"
        }
    } else {
        $HVMBin = hvm status --printExecPath 2>$null
        if ($HVMBin) {
            hvm use --useVersionInDotFile
            if ($LASTEXITCODE -ne 0) { return }
        } else {
            $HugoCommand = Get-Command -Name hugo.exe -ErrorAction SilentlyContinue -CommandType Application
            if ($HugoCommand) {
                $HVMBin = $HugoCommand.Definition
            }
        }
    }
    if ($HVMBin) {
        & $HVMBin @args
    } else {
        $global:LASTEXITCODE = 1
        Write-Error "Command not found." -ErrorAction Continue
    }
}
Set-Alias -Name hugo -Value Hugo-Override -Force
