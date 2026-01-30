<script>
    import { onMount, onDestroy } from "svelte";

    /** @type {string} */
    export let instanceId;
    export let status = "Offline";

    /** @type {string[]} */
    let logs = [];
    let command = "";
    /** @type {WebSocket | undefined} */
    let socket;
    /** @type {HTMLDivElement} */
    let consoleDiv;
    let connected = false;

    // Connect to WebSocket
    function connect() {
        if (!instanceId) return;

        // Determine Protocol
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        // Construct URL: /ws/instances/:id/console
        const wsUrl = `${protocol}//${window.location.host}/ws/instances/${instanceId}/console`;

        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            connected = true;
        };

        socket.onmessage = (/** @type {MessageEvent} */ event) => {
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

    // Connect on mount
    onMount(() => {
        connect();
    });

    // Reconnect if instanceId changes
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

        // Optimistic update
        // logs = [...logs, `> ${command}`];

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

    /** @param {KeyboardEvent} e */
    function handleKeydown(e) {
        if (e.key === "Enter") {
            sendCommand();
        }
    }
</script>

<div
    class="flex flex-col h-full bg-black/50 backdrop-blur-sm rounded-xl overflow-hidden border border-white/10 shadow-2xl font-mono text-sm group"
>
    <!-- Header/Title -->
    <div
        class="flex justify-between items-center px-4 py-2 bg-white/5 border-b border-white/5"
    >
        <div class="flex items-center gap-2">
            <div class="flex gap-1.5">
                <div class="w-3 h-3 rounded-full bg-red-500/50" />
                <div class="w-3 h-3 rounded-full bg-yellow-500/50" />
                <div class="w-3 h-3 rounded-full bg-green-500/50" />
            </div>
            <span
                class="ml-2 text-xs font-bold text-gray-500 uppercase tracking-widest"
                >Target: server.jar</span
            >
        </div>
        <div class="flex items-center gap-2">
            <span
                class="w-2 h-2 rounded-full {connected
                    ? 'bg-emerald-500 animate-pulse'
                    : 'bg-red-500'}"
            />
            <span class="text-xs text-gray-500 uppercase font-bold"
                >{connected ? "Live" : "Offline"}</span
            >
        </div>
    </div>

    <!-- Logs -->
    <div
        bind:this={consoleDiv}
        class="flex-1 overflow-y-auto p-4 space-y-0.5 scrollbar-thin scrollbar-thumb-gray-800 scrollbar-track-transparent"
    >
        {#if logs.length === 0}
            <div class="text-gray-600 italic">Waiting for logs...</div>
        {/if}
        {#each logs as log}
            <div
                class="text-gray-300 break-words font-medium leading-relaxed tracking-tight opacity-90 hover:opacity-100 transition-opacity"
            >
                <span class="text-gray-600 mr-2 select-none">$</span>{log}
            </div>
        {/each}
    </div>

    <!-- Input -->
    <div class="p-3 bg-white/5 border-t border-white/5">
        <div
            class="flex items-center gap-3 bg-black/20 rounded-lg px-3 py-2 border border-white/5 focus-within:border-indigo-500/50 focus-within:ring-1 focus-within:ring-indigo-500/50 transition-all duration-200"
        >
            <span class="text-indigo-400 font-bold">❯</span>
            <input
                type="text"
                bind:value={command}
                on:keydown={handleKeydown}
                class="flex-1 bg-transparent border-none text-gray-200 focus:ring-0 focus:outline-none placeholder-gray-600 text-sm font-medium"
                placeholder="Execute server command..."
            />
        </div>
    </div>
</div>
