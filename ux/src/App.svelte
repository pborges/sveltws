<script lang="ts">
    import Person from "./Person.svelte";

    let show: boolean = false;
</script>

<main>
    {#if show}
        <Person first="Bob" last="Jones"/>
        <Person first="Juliet" last="Fredrick"/>
        <Person first="Bob" last="Jones"/>
    {/if}
    <button on:click={() => (show = !show)}>{!show ? "SHOW" : "HIDE"}</button>
</main>
