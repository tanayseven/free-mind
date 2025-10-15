<script lang="ts">
	import Header from './Header.svelte';
	import '../app.css';

	let { children } = $props();
</script>

<div class="app min-h-screen flex flex-col">
	<Header />

	<main class="flex-1 flex flex-col items-center justify-center p-4 md:p-6 lg:p-8">
		{@render children()}
	</main>

	<footer class="w-full bg-secondary py-4 px-6 text-center text-secondary-foreground text-sm md:text-base">
		<p class="max-w-4xl mx-auto">
			Made with ❤️ and 😓 by Tanay PrabhuDesai visit <a href="https://freemind.tanay.tech" class="text-primary hover:underline">Freemind</a> to learn about this application.
		</p>
	</footer>
</div>

<style>
	/* Additional styles can be added here if needed */
</style>
