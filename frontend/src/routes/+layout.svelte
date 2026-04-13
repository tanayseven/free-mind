<script lang="ts">
	import Footer from './Footer.svelte';
	import '../app.css';

	let { children } = $props();
</script>

<div class="app flex h-screen flex-col overflow-hidden">
	<main class="flex min-h-0 flex-1 flex-col">
		{@render children()}
	</main>

	<Footer />
</div>

<style>
	/* Additional styles can be added here if needed */
</style>
