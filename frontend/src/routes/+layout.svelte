<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<div class="app flex h-screen flex-col overflow-hidden">
	<main class="flex min-h-0 flex-1 flex-col">
		{@render children()}
	</main>
</div>
