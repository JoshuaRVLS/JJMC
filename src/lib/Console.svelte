<script>
    import { onMount, onDestroy } from "svelte";

    export let instanceId;

    let logs = [];
    let command = "";

    let socket;

    let consoleDiv;
    let connected = false;

    function connect() {
        if (!instanceId) return;

        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";

        const wsUrl = `${protocol}//${window.location.host}/ws/instances/${instanceId}/console`;

        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            connected = true;
        };

        socket.onmessage = (event) => {
            logs = [...logs, event.data];
            scrollToBottom();
        };

        socket.onclose = () => {
            connected = false;
            logs = [...logs, "Disconnected from server..."];
        };

        socket.onerror = () => {
            connected = false;
        };
    }

    onMount(() => {
        connect();
    });

    $: if (instanceId) {
        if (socket) socket.close();
        connect();
    }

    onDestroy(() => {
        if (socket) socket.close();
    });

    function scrollToBottom() {
        if (consoleDiv) {
            setTimeout(() => {
                consoleDiv.scrollTop = consoleDiv.scrollHeight;
            }, 0);
        }
    }

    async function sendCommand() {
        if (!command.trim()) return;

        try {
            await fetch(`/api/instances/${instanceId}/command`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ command }),
            });
            command = "";
        } catch (e) {
            console.error("Failed to send command", e);
        }
    }

    function handleKeydown(e) {
        if (e.key === "Enter") {
            sendCommand();
        }
    }
    function handleConsoleClick(event) {
        const link = event.target.closest("a.console-link");
        if (link) {
            if (!event.ctrlKey && !event.metaKey) {
                event.preventDefault();
            }
        }
    }

    function formatLog(text) {
        if (!text) return "";

        text = text
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;");

        const urlRegex = /(https?:\/\/[^\s]+)/g;
        text = text.replace(
            urlRegex,
            '<a href="$1" target="_blank" rel="noopener noreferrer" class="console-link underline decoration-indigo-500/50 hover:decoration-indigo-400 text-indigo-300 hover:text-indigo-200 transition-colors cursor-pointer" title="Ctrl + Click to open">$1</a>',
        );

        const ansiRegex = /\x1b\[([0-9;]*)m/g;

        let html = "";
        let lastIndex = 0;
        let styleStack = [];
        let match;

        while ((match = ansiRegex.exec(text)) !== null) {
            html += text.substring(lastIndex, match.index);
            lastIndex = ansiRegex.lastIndex;

            const codes = match[1].split(";").map(Number);

            for (let i = 0; i < codes.length; i++) {
                const code = codes[i];

                if (code === 0) {
                    while (styleStack.length > 0) {
                        html += "</span>";
                        styleStack.pop();
                    }
                } else if (code === 1) {
                    html += "<span class='font-bold'>";
                    styleStack.push("span");
                } else if (code === 38 && codes[i + 1] === 2) {
                    const r = codes[i + 2];
                    const g = codes[i + 3];
                    const b = codes[i + 4];
                    html += `<span style="color: rgb(${r},${g},${b})">`;
                    styleStack.push("span");
                    i += 4;
                } else if (code === 48 && codes[i + 1] === 2) {
                    const r = codes[i + 2];
                    const g = codes[i + 3];
                    const b = codes[i + 4];
                    html += `<span style="background-color: rgb(${r},${g},${b})">`;
                    styleStack.push("span");
                    i += 4;
                } else if (code >= 30 && code <= 37) {
                    const colors = [
                        "black",
                        "red",
                        "green",
                        "yellow",
                        "blue",
                        "magenta",
                        "cyan",
                        "white",
                        "bright-black",
                    ];

                    const map = [
                        "#000000",
                        "#ef4444",
                        "#22c55e",
                        "#eab308",
                        "#3b82f6",
                        "#d946ef",
                        "#06b6d4",
                        "#f1f5f9",
                    ];
                    html += `<span style="color: ${map[code - 30]}">`;
                    styleStack.push("span");
                } else if (code >= 90 && code <= 97) {
                    const map = [
                        "#6b7280",
                        "#f87171",
                        "#4ade80",
                        "#facc15",
                        "#60a5fa",
                        "#e879f9",
                        "#22d3ee",
                        "#ffffff",
                    ];
                    html += `<span style="color: ${map[code - 90]}">`;
                    styleStack.push("span");
                }
            }
        }

        html += text.substring(lastIndex);

        while (styleStack.length > 0) {
            html += "</span>";
            styleStack.pop();
        }

        return html;
    }
</script>

<div
    class="flex flex-col h-full bg-slate-950/80 backdrop-blur-xl rounded-2xl overflow-hidden border border-white/5 shadow-2xl font-mono text-sm group relative"
>
    <div
        class="absolute inset-0 pointer-events-none bg-linear-to-tr from-indigo-500/5 via-transparent to-emerald-500/5 opacity-50"
    ></div>

    <div
        class="flex justify-between items-center px-5 py-3 bg-white/5 border-b border-white/5 relative z-10"
    >
        <div class="flex items-center gap-3">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="w-4 h-4 text-gray-400"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><polyline points="4 17 10 11 4 5"></polyline><line
                    x1="12"
                    y1="19"
                    x2="20"
                    y2="19"
                ></line></svg
            >
            <span
                class="text-xs font-semibold text-gray-400 uppercase tracking-wider"
                >Server Console</span
            >
        </div>
        <div class="flex items-center gap-2">
            <div class="relative flex h-2 w-2">
                {#if connected}
                    <span
                        class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
                    ></span>
                    <span
                        class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"
                    ></span>
                {:else}
                    <span
                        class="relative inline-flex rounded-full h-2 w-2 bg-rose-500"
                    ></span>
                {/if}
            </div>
            <span
                class="text-[10px] font-bold tracking-widest uppercase {connected
                    ? 'text-emerald-500'
                    : 'text-rose-500'}">{connected ? "Online" : "Offline"}</span
            >
        </div>
    </div>

    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
        bind:this={consoleDiv}
        on:click={handleConsoleClick}
        role="log"
        class="flex-1 overflow-y-auto p-5 space-y-1 scrollbar-thin scrollbar-thumb-white/10 scrollbar-track-transparent relative z-10"
    >
        {#if logs.length === 0}
            <div
                class="flex flex-col items-center justify-center h-full text-gray-700 space-y-2"
            >
                <div class="text-4xl opacity-20">_</div>
                <div class="text-xs uppercase tracking-widest opacity-50">
                    No output logs
                </div>
            </div>
        {/if}
        {#each logs as log}
            <div
                class="break-words font-medium leading-relaxed tracking-tight text-slate-300/90 hover:text-white transition-colors duration-150 animate-in fade-in slide-in-from-bottom-1"
            >
                <span class="text-indigo-400/50 mr-3 select-none text-xs"
                    >➜</span
                >{@html formatLog(log)}
            </div>
        {/each}
    </div>

    <div class="p-4 bg-white/2 border-t border-white/5 relative z-10">
        <div
            class="group/input flex items-center gap-3 bg-black/40 rounded-xl px-4 py-3 border border-white/5 focus-within:border-indigo-500/50 focus-within:ring-2 focus-within:ring-indigo-500/20 transition-all duration-300 shadow-inner"
        >
            <span class="text-indigo-400 font-bold animate-pulse">❯</span>
            <input
                type="text"
                bind:value={command}
                on:keydown={handleKeydown}
                class="flex-1 bg-transparent border-none text-gray-100 placeholder-gray-600 focus:ring-0 focus:outline-none text-sm font-medium tracking-wide"
                placeholder="Enter command..."
            />
            <div
                class="text-[10px] text-gray-600 font-medium px-2 py-0.5 rounded border border-white/5 hidden group-focus-within/input:block animate-in fade-in zoom-in-95"
            >
                ENTER
            </div>
        </div>
    </div>
</div>
