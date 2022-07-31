<script lang="ts">
    export let first: string;
    export let last: string;

    import people from "./peoplestore";

    const person = people.bind({first, last});

    const replacer = (k: string, v: any) => {
        if (v instanceof Error) {
            return v.message
        }
        return v
    }

    let error: Error | undefined;
</script>

<main>
    <div>
        <pre>{JSON.stringify($person, replacer, "  ")}</pre>
    </div>
    <button on:click={()=>{
        $person.value.reset().catch(e=>error = e)
    }}>RESET
    </button>
    {#if error}
        {error.message}
    {/if}
</main>
