<script>
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";

    onMount(() => {
        goto("/instances");
    });
</script>

<div class="p-10 text-center font-mono animate-pulse">
    Redirecting to Instances...
</div>
