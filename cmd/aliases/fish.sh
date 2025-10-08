# Hugo Version Manager: override path to the hugo executable.
function hugo --wraps hugo --description 'Hugo Version Manager: override path to the hugo executable.'
    set -l hvm_show_status true
    set -l hugo_bin

    if set hugo_bin (hvm status --printExecPathCached)
        if test "$hvm_show_status" = "true"
            printf "Hugo version management is enabled in this directory.\n" ^&2
            printf "Run 'hvm status' for details, or 'hvm disable' to disable.\n\n" ^&2
        end
    else if set hugo_bin (hvm status --printExecPath)
        if not hvm use --useVersionInDotFile
            return 1
        end
    else
        if not set hugo_bin (which hugo)
            printf "Command not found.\n" ^&2
            return 1
        end
    end
    "$hugo_bin" $argv
end
