<script>
    import { onMount, onDestroy } from "svelte";
    import Chart from "chart.js/auto";

    export let instanceId;

    let cpuCanvas;
    let memoryCanvas;
    let cpuChart;
    let memoryChart;

    let socket;
    let connected = false;

    // Config options for charts to look "premium"
    const commonOptions = {
        responsive: true,
        maintainAspectRatio: false,
        animation: false,
        plugins: {
            legend: {
                display: false,
            },
            tooltip: {
                mode: "index",
                intersect: false,
                backgroundColor: "rgba(0,0,0,0.8)",
                titleColor: "#fff",
                bodyColor: "#fff",
                borderColor: "rgba(255,255,255,0.1)",
                borderWidth: 1,
            },
        },
        scales: {
            x: {
                display: false, // Hide time axis for clean look
                grid: {
                    display: false,
                },
            },
            y: {
                display: true,
                position: "right",
                grid: {
                    color: "rgba(255,255,255,0.05)",
                },
                ticks: {
                    color: "rgba(255,255,255,0.3)",
                    font: {
                        size: 9,
                    },
                },
                min: 0,
            },
        },
        elements: {
            point: {
                radius: 0, // Hide points
                hitRadius: 10,
            },
            line: {
                tension: 0.4, // Smooth curves
                borderWidth: 2,
            },
        },
    };

    function initCharts() {
        if (cpuCanvas) {
            cpuChart = new Chart(cpuCanvas, {
                type: "line",
                data: {
                    labels: [],
                    datasets: [
                        {
                            label: "CPU %",
                            data: [],
                            borderColor: "#818cf8", // Indigo 400
                            backgroundColor: (context) => {
                                const ctx = context.chart.ctx;
                                const gradient = ctx.createLinearGradient(
                                    0,
                                    0,
                                    0,
                                    200,
                                );
                                gradient.addColorStop(
                                    0,
                                    "rgba(129, 140, 248, 0.5)",
                                ); // Indigo
                                gradient.addColorStop(
                                    1,
                                    "rgba(129, 140, 248, 0)",
                                );
                                return gradient;
                            },
                            fill: true,
                        },
                    ],
                },
                options: {
                    ...commonOptions,
                    scales: {
                        ...commonOptions.scales,
                        y: { ...commonOptions.scales.y, max: 100 },
                    },
                },
            });
        }

        if (memoryCanvas) {
            memoryChart = new Chart(memoryCanvas, {
                type: "line",
                data: {
                    labels: [],
                    datasets: [
                        {
                            label: "Memory (MB)",
                            data: [],
                            borderColor: "#34d399", // Emerald 400
                            backgroundColor: (context) => {
                                const ctx = context.chart.ctx;
                                const gradient = ctx.createLinearGradient(
                                    0,
                                    0,
                                    0,
                                    200,
                                );
                                gradient.addColorStop(
                                    0,
                                    "rgba(52, 211, 153, 0.5)",
                                ); // Emerald
                                gradient.addColorStop(
                                    1,
                                    "rgba(52, 211, 153, 0)",
                                );
                                return gradient;
                            },
                            fill: true,
                        },
                    ],
                },
                options: commonOptions,
            });
        }
    }

    function updateChart(chart, value, label) {
        if (!chart) return;

        const maxPoints = 50;
        chart.data.labels.push(label);
        chart.data.datasets[0].data.push(value);

        if (chart.data.labels.length > maxPoints) {
            chart.data.labels.shift();
            chart.data.datasets[0].data.shift();
        }

        chart.update("none"); // 'none' for performance improvement by skipping animations for updates
    }

    function connect() {
        if (!instanceId) return;

        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        const wsUrl = `${protocol}//${window.location.host}/ws/instances/${instanceId}/stats`;

        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            connected = true;
        };

        socket.onmessage = (event) => {
            try {
                const stats = JSON.parse(event.data);
                const timeLabel = new Date(
                    stats.time * 1000,
                ).toLocaleTimeString();

                if (cpuChart) updateChart(cpuChart, stats.cpu, timeLabel);
                if (memoryChart)
                    updateChart(
                        memoryChart,
                        stats.memory / 1024 / 1024,
                        timeLabel,
                    ); // Convert bytes to MB
            } catch (e) {
                console.error("Failed to parse stats", e);
            }
        };

        socket.onclose = () => {
            connected = false;
        };
    }

    onMount(() => {
        initCharts();
        connect();
    });

    onDestroy(() => {
        if (socket) socket.close();
        if (cpuChart) cpuChart.destroy();
        if (memoryChart) memoryChart.destroy();
    });
</script>

<div class="grid grid-cols-2 gap-4 h-full">
    <div
        class="bg-gray-900/50 border border-white/5 rounded-2xl p-4 flex flex-col relative overflow-hidden"
    >
        <div class="flex justify-between items-center mb-2 z-10 relative">
            <h3
                class="text-xs font-bold text-gray-400 uppercase tracking-widest"
            >
                CPU Usage
            </h3>
            <span class="text-indigo-400 font-mono text-xs font-bold">
                {cpuChart?.data?.datasets[0]?.data.slice(-1)[0]?.toFixed(1) ||
                    0}%
            </span>
        </div>
        <div class="flex-1 min-h-0 relative z-10">
            <canvas bind:this={cpuCanvas}></canvas>
        </div>
        <!-- Background Glow -->
        <div
            class="absolute bottom-[-50%] right-[-20%] w-32 h-32 bg-indigo-500/20 blur-3xl rounded-full"
        ></div>
    </div>

    <div
        class="bg-gray-900/50 border border-white/5 rounded-2xl p-4 flex flex-col relative overflow-hidden"
    >
        <div class="flex justify-between items-center mb-2 z-10 relative">
            <h3
                class="text-xs font-bold text-gray-400 uppercase tracking-widest"
            >
                Memory
            </h3>
            <span class="text-emerald-400 font-mono text-xs font-bold">
                {memoryChart?.data?.datasets[0]?.data
                    .slice(-1)[0]
                    ?.toFixed(0) || 0} MB
            </span>
        </div>
        <div class="flex-1 min-h-0 relative z-10">
            <canvas bind:this={memoryCanvas}></canvas>
        </div>
        <!-- Background Glow -->
        <div
            class="absolute bottom-[-50%] right-[-20%] w-32 h-32 bg-emerald-500/20 blur-3xl rounded-full"
        ></div>
    </div>
</div>
