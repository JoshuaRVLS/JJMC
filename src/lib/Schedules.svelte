<script>
    import { onMount } from "svelte";
    import {
        Loader2,
        Plus,
        Trash2,
        Calendar,
        Play,
        Power,
        Database,
    } from "lucide-svelte";
    import { addToast } from "$lib/stores/toast";

    export let instanceId;

    let schedules = [];
    let loading = true;
    let isModalOpen = false;

    let newSchedule = {
        name: "",
        cronExpression: "0 0 * * *",
        type: "command",
        payload: "",
        enabled: true,
    };

    async function loadSchedules() {
        loading = true;
        try {
            const res = await fetch(`/api/instances/${instanceId}/schedules`);
            if (res.ok) {
                schedules = await res.json();
            } else {
                addToast("Failed to load schedules", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error loading schedules", "error");
        } finally {
            loading = false;
        }
    }

    async function createSchedule() {
        try {
            const res = await fetch(`/api/instances/${instanceId}/schedules`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(newSchedule),
            });

            if (res.ok) {
                addToast("Schedule created", "success");
                isModalOpen = false;
                loadSchedules();
                newSchedule = {
                    name: "",
                    cronExpression: "0 0 * * *",
                    type: "command",
                    payload: "",
                    enabled: true,
                };
            } else {
                const data = await res.json();
                addToast(data.error || "Failed to create schedule", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error creating schedule", "error");
        }
    }

    async function deleteSchedule(id) {
        if (!confirm("Are you sure you want to delete this schedule?")) return;

        try {
            const res = await fetch(
                `/api/instances/${instanceId}/schedules/${id}`,
                {
                    method: "DELETE",
                },
            );

            if (res.ok) {
                addToast("Schedule deleted", "success");
                loadSchedules();
            } else {
                addToast("Failed to delete schedule", "error");
            }
        } catch (e) {
            console.error(e);
            addToast("Error deleting schedule", "error");
        }
    }

    onMount(() => {
        loadSchedules();
    });

    const getTypeIcon = (type) => {
        switch (type) {
            case "restart":
                return Power;
            case "backup":
                return Database;
            default:
                return Calendar;
        }
    };
</script>

<div class="h-full flex flex-col gap-6">
    <div class="flex justify-between items-center">
        <div>
            <h2 class="text-xl font-bold text-white">Scheduled Tasks</h2>
            <p class="text-gray-400 text-sm">Automate server operations</p>
        </div>
        <button
            on:click={() => (isModalOpen = true)}
            class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors font-medium shadow-lg shadow-indigo-900/20"
        >
            <Plus class="w-4 h-4" />
            New Schedule
        </button>
    </div>

    {#if loading}
        <div class="flex-1 flex items-center justify-center">
            <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
        </div>
    {:else if schedules.length === 0}
        <div
            class="flex-1 flex flex-col items-center justify-center text-gray-500 gap-4"
        >
            <Calendar class="w-16 h-16 opacity-20" />
            <p>No scheduled tasks found</p>
        </div>
    {:else}
        <div
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 overflow-y-auto min-h-0"
        >
            {#each schedules as schedule}
                <div
                    class="bg-gray-900/50 border border-white/5 p-4 rounded-xl flex flex-col gap-4 group hover:border-indigo-500/30 transition-colors"
                >
                    <div class="flex justify-between items-start">
                        <div class="flex items-center gap-3">
                            <div
                                class="w-10 h-10 rounded-full bg-indigo-500/10 flex items-center justify-center text-indigo-400"
                            >
                                <svelte:component
                                    this={getTypeIcon(schedule.type)}
                                    class="w-5 h-5"
                                />
                            </div>
                            <div>
                                <h3 class="font-bold text-white">
                                    {schedule.name}
                                </h3>
                                <div
                                    class="text-xs text-mono text-indigo-300 bg-indigo-500/10 px-2 py-0.5 rounded-full inline-block mt-1"
                                >
                                    {schedule.cronExpression}
                                </div>
                            </div>
                        </div>
                        <button
                            on:click={() => deleteSchedule(schedule.id)}
                            class="text-gray-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                            <Trash2 class="w-4 h-4" />
                        </button>
                    </div>

                    <div
                        class="bg-black/20 p-3 rounded-lg text-sm font-mono text-gray-300 break-all"
                    >
                        {#if schedule.type === "command"}
                            <span class="text-gray-500 select-none">$</span>
                            {schedule.payload}
                        {:else}
                            <span
                                class="text-indigo-400 uppercase text-xs font-bold"
                                >{schedule.type}</span
                            >
                        {/if}
                    </div>

                    <div
                        class="mt-auto flex justify-between items-center text-xs text-gray-500 pt-2 border-t border-white/5"
                    >
                        <span
                            >Last run: {schedule.lastRun
                                ? new Date(
                                      schedule.lastRun * 1000,
                                  ).toLocaleString()
                                : "Never"}</span
                        >
                        <span
                            class={schedule.enabled
                                ? "text-green-400"
                                : "text-red-400"}
                        >
                            {schedule.enabled ? "Active" : "Disabled"}
                        </span>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>

{#if isModalOpen}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
        <div
            class="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md shadow-2xl p-6"
        >
            <h3 class="text-xl font-bold text-white mb-6">Create Schedule</h3>

            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-400 mb-1"
                        >Name</label
                    >
                    <input
                        type="text"
                        bind:value={newSchedule.name}
                        placeholder="Daily Restart"
                        class="w-full bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-indigo-500 transition-colors"
                    />
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-400 mb-1"
                        >Cron Expression</label
                    >
                    <div class="flex gap-2">
                        <input
                            type="text"
                            bind:value={newSchedule.cronExpression}
                            placeholder="0 0 * * *"
                            class="flex-1 bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white font-mono focus:outline-none focus:border-indigo-500 transition-colors"
                        />
                        <a
                            href="https://crontab.guru/"
                            target="_blank"
                            class="px-3 py-2 bg-white/5 hover:bg-white/10 rounded-lg text-gray-400 transition-colors flex items-center"
                            >?</a
                        >
                    </div>
                    <p class="text-xs text-gray-500 mt-1">
                        Standard cron syntax (min hour dom month dow)
                    </p>
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-400 mb-1"
                        >Type</label
                    >
                    <select
                        bind:value={newSchedule.type}
                        class="w-full bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-indigo-500 transition-colors"
                    >
                        <option value="command">Console Command</option>
                        <option value="restart">Restart Server</option>
                        <option value="start">Start Server</option>
                        <option value="stop">Stop Server</option>
                        <!-- <option value="backup">Create Backup</option> -->
                    </select>
                </div>

                {#if newSchedule.type === "command"}
                    <div>
                        <label
                            class="block text-sm font-medium text-gray-400 mb-1"
                            >Command</label
                        >
                        <input
                            type="text"
                            bind:value={newSchedule.payload}
                            placeholder="say Hello World"
                            class="w-full bg-black/20 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-indigo-500 transition-colors"
                        />
                    </div>
                {/if}
            </div>

            <div class="flex justify-end gap-3 mt-8">
                <button
                    on:click={() => (isModalOpen = false)}
                    class="px-4 py-2 text-gray-400 hover:text-white transition-colors"
                >
                    Cancel
                </button>
                <button
                    on:click={createSchedule}
                    disabled={!newSchedule.name || !newSchedule.cronExpression}
                    class="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    Create Schedule
                </button>
            </div>
        </div>
    </div>
{/if}
