<script>
    import { slide } from "svelte/transition";
    import { createEventDispatcher } from "svelte";

    export let value;

    export let options = [];
    export let label = "";
    export let placeholder = "Select an option";
    export let id = "select-" + Math.random().toString(36).slice(2);
    export let className = "";

    const dispatch = createEventDispatcher();
    let isOpen = false;

    $: selectedOption = options.find((o) =>
        typeof o === "object" ? o.value === value : o === value,
    );
    $: displayValue = selectedOption
        ? typeof selectedOption === "object"
            ? selectedOption.label
            : selectedOption
        : placeholder;

    function toggle() {
        isOpen = !isOpen;
    }

    // ... (rest of functions)

    // NOTE: Removed duplicate functions for brevity in prompt, but replacing the script section carefully.
    // Wait, replacing the whole script is safer or targeted. Targeted is better.
    // I need to split this into two replaces ideally or one big one.
    // I'll replace the export section and the label/button section.
    // Wait, `replace_file_content` is single contiguous block.
    // I'll replace the exports first.

    function select(option) {
        value = typeof option === "object" ? option.value : option;
        isOpen = false;
        dispatch("change", value);
    }

    function clickOutside(node) {
        const handleClick = (event) => {
            if (
                node &&
                !node.contains(event.target) &&
                !event.defaultPrevented
            ) {
                node.dispatchEvent(new CustomEvent("click_outside"));
            }
        };

        document.addEventListener("click", handleClick, true);

        return {
            destroy() {
                document.removeEventListener("click", handleClick, true);
            },
        };
    }
</script>

<div
    class="relative"
    use:clickOutside
    on:click_outside={() => (isOpen = false)}
>
    {#if label}
        <label
            for={id}
            class="block text-xs font-bold uppercase text-gray-400 mb-2 ml-1"
            >{label}</label
        >
    {/if}

    <button
        type="button"
        {id}
        on:click={toggle}
        class="w-full bg-black/20 border border-white/10 rounded-xl p-4 text-left text-white flex justify-between items-center transition-all hover:bg-black/30 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 {className}"
    >
        <span class={!selectedOption ? "text-gray-500" : ""}
            >{displayValue}</span
        >
        <svg
            class="w-4 h-4 text-gray-500 transition-transform duration-200 {isOpen
                ? 'rotate-180'
                : ''}"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 9l-7 7-7-7"
            />
        </svg>
    </button>

    {#if isOpen}
        <div
            class="absolute z-50 w-full mt-2 bg-gray-900 border border-white/10 rounded-xl shadow-2xl overflow-hidden backdrop-blur-xl"
            transition:slide={{ duration: 200 }}
        >
            <div class="max-h-60 overflow-y-auto py-1">
                {#each options as option}
                    <button
                        type="button"
                        class="w-full text-left px-4 py-3 text-sm text-gray-300 hover:bg-indigo-500/20 hover:text-white transition-colors flex items-center justify-between group"
                        on:click={() => select(option)}
                    >
                        <span
                            >{typeof option === "object"
                                ? option.label
                                : option}</span
                        >
                        {#if (typeof option === "object" ? option.value : option) === value}
                            <svg
                                class="w-4 h-4 text-indigo-400"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M5 13l4 4L19 7"
                                />
                            </svg>
                        {/if}
                    </button>
                {/each}
            </div>
        </div>
    {/if}
</div>
