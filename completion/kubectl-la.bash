#!/usr/bin/env bash

_kubectl_label_assist_completions() {
    local cur prev opts labels resource_type namespace all_namespaces

    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Determine the resource type based on the word following 'get'
    resource_type=""
    namespace=""
    all_namespaces="false"
    for i in "${!COMP_WORDS[@]}"; do
        if [[ "${COMP_WORDS[i]}" == "get" && $((i + 1)) -lt ${#COMP_WORDS[@]} ]]; then
            resource_type="${COMP_WORDS[$((i + 1))]}"
        elif [[ "${COMP_WORDS[i]}" == "-n" && $((i + 1)) -lt ${#COMP_WORDS[@]} ]]; then
            namespace="${COMP_WORDS[$((i + 1))]}"
        elif [[ "${COMP_WORDS[i]}" == "-A" ]]; then
            all_namespaces="true"
        fi
    done

    # Default to pods if no resource type is provided
    resource_type="${resource_type:-pods}"

    # Set namespace to "all" if -A is provided
    if [[ "$all_namespaces" == "true" ]]; then
        namespace="all"
    else
        # Default to "default" namespace if not provided
        namespace="${namespace:-default}"
    fi

    # Fetch dynamic labels from the command `./kl --resource <resource_type> --namespace <namespace>`
    labels=$(kubectl-la-autocomplete --resource "${resource_type}" --namespace "${namespace}" 2>/dev/null)

    # If the command fails, return early to avoid breaking the autocomplete
    if [ $? -ne 0 ]; then
        return
    fi

    # Handle label completion
    if [[ "${prev}" == "-l" ]]; then
        if [[ "${cur}" == *"="* ]]; then
            local label="${cur%%=*}"
            local value="${cur#*=}"

            # Get available values for the label
            opts=$(echo "${labels}" | awk -F= -v label="$label" '$1 == label {print $2}' | sort -u)
            # Generate completions based on partial match
            COMPREPLY=($(compgen -W "${opts}" -X "!*${value}*" -- "${value}"))
        else
            # Get available label keys
            opts=$(echo "${labels}" | awk -F= '{print $1}' | sort -u)
            # Generate completions based on partial match
            COMPREPLY=($(compgen -W "${opts}" -X "!*${cur}*" -- "${cur}"))
        fi
    fi
}

complete -F _kubectl_label_assist_completions kubectl