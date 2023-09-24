# Hugo Version Manager: override path to the hugo executable.
function Hugo-Override {
  Set-Variable -Name "hvm_show_status" -Value $true
  Set-Variable -Name "hugo_bin" -Value $(hvm status --printExecPathCached)
  If ($hugo_bin) {
    If ($hvm_show_status) {
      echo "Hugo version management is enabled in this directory."
      echo "Run 'hvm status' for details, or 'hvm disable' to disable.`n"
    }
    & "$hugo_bin" $args
  } Else {
    Set-Variable -Name "hugo_bin" -Value $(hvm status --printExecPath)
    If ($hugo_bin) {
      hvm use --useVersionInDotFile
      if ($lastexitcode) {
        return
      }
    } Else {
      Set-Variable -Name "hugo_bin" -Value $((gcm hugo.exe).Path 2> $null)
      If ($hugo_bin) {
        & "$hugo_bin" $args
      } Else {
        echo "Command not found"
      }
    }
  }
}
Set-Alias hugo Hugo-Override
