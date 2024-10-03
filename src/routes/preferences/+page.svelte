<script>
import {Button} from "$lib/components/ui/button/index.js";
import {goto} from "$app/navigation";
import {ArrowBigLeft} from "lucide-svelte";
</script>

<Button variant="outline" class="text-xl" on:click={()=>goto('/')}>
        <ArrowBigLeft class="mr-2 h-6 w-6" />
        Back
</Button>