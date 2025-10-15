<script lang="ts">
	import Header from './Header.svelte';
	import '../app.css';

	let { children } = $props();
</script>

<div class="app">
	<Header />

	<main>
		{@render children()}
	</main>

	<footer>
		<p>
			Made with ❤️ and 😓 by Tanay PrabhuDesai visit <a href="https://freemind.tanay.tech">Freemind</a> to learn about this application.
		</p>
	</footer>
</div>

<style>

</style>
