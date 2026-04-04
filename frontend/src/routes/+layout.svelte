<script lang="ts">
	import Header from './Header.svelte';
	import Footer from './Footer.svelte';
	import '../app.css';

	let { children } = $props();
</script>

<div class="app min-h-screen flex flex-col">
	<Header />

	<main class="flex-1 flex flex-col items-center justify-center p-4 md:p-6 lg:p-8">
		{@render children()}
	</main>

	<Footer />
</div>

<style>
	/* Additional styles can be added here if needed */
</style>
