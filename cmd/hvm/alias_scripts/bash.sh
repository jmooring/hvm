# Hugo Version Manager: override path to the hugo executable.
hugo() {
  hvm_show_status=true
  if hugo_bin=$(hvm status --printExecPath); then
    if [ "${hvm_show_status}" == true ]; then
      >&2 printf "Hugo version management is enabled in this directory.\\n"
      >&2 printf "Run 'hvm status' for details, or 'hvm disable' to disable.\\n\\n"
    fi
  else
    if ! hugo_bin=$(which hugo); then
      >&2 printf "Command not found.\\n"
      return 1
    fi
  fi
  "${hugo_bin}" "$@"
}
