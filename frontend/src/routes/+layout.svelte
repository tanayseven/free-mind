<script lang="ts">
	import Footer from './Footer.svelte';
	import '../app.css';

	let { children } = $props();
</script>

<div class="app min-h-screen flex flex-col">
	<main class="flex-1 flex flex-col">
		{@render children()}
	</main>

	<Footer />
</div>

<style>
	/* Additional styles can be added here if needed */
</style>
