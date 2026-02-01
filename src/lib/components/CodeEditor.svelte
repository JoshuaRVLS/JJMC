<script>
    import { onMount } from "svelte";

     
    export let value = "";
     
    export let language = "json";
     
    export let readonly = false;

     
    let preElement;
     
    let textareaElement;

     
     
    function highlight(code) {
        if (!code) return "";

        let html = code
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;");

        if (language === "json") {
            html = html.replace(
                /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
                /** @param {string} match */
                (match) => {
                    let cls = "text-indigo-400"; // number
                    if (/^"/.test(match)) {
                        if (/:$/.test(match)) {
                            cls = "text-sky-400 font-bold";  
                        } else {
                            cls = "text-emerald-400";  
                        }
                    } else if (/true|false/.test(match)) {
                        cls = "text-rose-400 font-bold";  
                    } else if (/null/.test(match)) {
                        cls = "text-gray-500 font-bold";  
                    }
                    return `<span class="${cls}">${match}</span>`;
                },
            );
        } else if (language === "toml" || language === "properties") {
             
             
             
             
             
             
             
             
             
             
            const tokenRegex =
                /("(\\.|[^"\\])*")|(#.*$)|(^\[.*\]$)|(^\s*[a-zA-Z0-9_\-.]+(?=\s*=))|(\b(true|false)\b)|(\b\d+\b)/gm;

            html = html.replace(
                tokenRegex,
                (
                    match,
                    str,
                    strInner,
                    comment,
                    section,
                    key,
                    bool,
                    boolInner,
                    num,
                ) => {
                    if (str)
                        return `<span class="text-emerald-400">${match}</span>`;
                    if (comment)
                        return `<span class="text-gray-500">${match}</span>`;
                    if (section)
                        return `<span class="text-yellow-400 font-bold">${match}</span>`;
                    if (key)
                        return `<span class="text-sky-400">${match}</span>`;
                    if (bool)
                        return `<span class="text-rose-400 font-bold">${match}</span>`;
                    if (num)
                        return `<span class="text-indigo-400">${match}</span>`;
                    return match;
                },
            );
        }

        return html;
    }

    function syncScroll() {
        if (preElement && textareaElement) {
            preElement.scrollTop = textareaElement.scrollTop;
            preElement.scrollLeft = textareaElement.scrollLeft;
        }
    }

    /** @param {Event} e */
    function handleInput(e) {
        const target = /** @type {HTMLTextAreaElement} */ (e.target);
        value = target.value;
    }

    // Handle special keys
    /** @param {KeyboardEvent} e */
    function handleKeydown(e) {
        const target = /** @type {HTMLTextAreaElement} */ (e.target);
        const start = target.selectionStart;
        const end = target.selectionEnd;

        if (e.key === "Tab") {
            e.preventDefault();
            value = value.substring(0, start) + "    " + value.substring(end);
            // Move cursor
            setTimeout(() => {
                target.selectionStart = target.selectionEnd = start + 4;
            }, 0);
            return;
        }

        if (e.key === "Enter") {
            e.preventDefault();
            const lineStart = value.lastIndexOf("\n", start - 1) + 1;
            const lineContent = value.substring(lineStart, start);
            const indentMatch = lineContent.match(/^\s*/);
            let indent = indentMatch ? indentMatch[0] : "";

            // Smart indent
            if (
                lineContent.trim().endsWith("{") ||
                lineContent.trim().endsWith("[")
            ) {
                indent += "    ";
            }

            value =
                value.substring(0, start) +
                "\n" +
                indent +
                value.substring(end);

            setTimeout(() => {
                target.selectionStart = target.selectionEnd =
                    start + 1 + indent.length;
            }, 0);
            return;
        }

        // Auto close brackets/quotes
        /** @type {Record<string, string>} */
        const pairs = {
            "(": ")",
            "[": "]",
            "{": "}",
            '"': '"',
            "'": "'",
        };

        if (pairs[e.key]) {
            e.preventDefault();
            value =
                value.substring(0, start) +
                e.key +
                pairs[e.key] +
                value.substring(end);
            setTimeout(() => {
                target.selectionStart = target.selectionEnd = start + 1;
            }, 0);
        }
    }
</script>

<div class="relative w-full h-full font-mono text-sm group">
    
    <pre
        bind:this={preElement}
        class="absolute inset-0 m-0 p-4 pointer-events-none overflow-hidden bg-[#0b0e14] text-gray-300 whitespace-pre-wrap break-words"
        aria-hidden="true">{@html highlight(value)}<br /></pre>

    
    <textarea
        bind:this={textareaElement}
        {value}
        on:input={handleInput}
        on:scroll={syncScroll}
        on:keydown={handleKeydown}
        {readonly}
        spellcheck="false"
        class="absolute inset-0 w-full h-full p-4 bg-transparent text-transparent caret-white focus:outline-none resize-none whitespace-pre-wrap break-words selection:bg-indigo-500/30 font-mono"
    ></textarea>
</div>
