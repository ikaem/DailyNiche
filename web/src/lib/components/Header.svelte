<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';

	// resolve() returns a relative URL (e.g. "./dashboard"), suitable for
	// href but not for comparing against page.url.pathname (always
	// absolute) - path is kept separately, just for that comparison.
	const links = [
		{ path: '/', href: resolve('/'), label: 'Home' },
		{ path: '/dashboard', href: resolve('/dashboard'), label: 'Dashboard' }
	];
</script>

<header class="site">
	<div class="logo">Daily<span>Niche</span></div>
	<nav>
		{#each links as link (link.path)}
			<a href={link.href} class:active={page.url.pathname === link.path}>{link.label}</a>
		{/each}
	</nav>
</header>

<style>
	header.site {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.25rem 2rem;
		background: var(--paper);
		border-bottom: 1px solid var(--line);
	}
	.logo {
		font-family: 'Playfair Display', Georgia, serif;
		font-size: 1.7rem;
		font-weight: 700;
		font-style: italic;
	}
	.logo span {
		color: var(--accent);
	}
	nav a {
		color: var(--ink);
		text-decoration: none;
		margin-left: 1.75rem;
		font-size: 0.9rem;
		font-weight: 500;
		padding-bottom: 0.2rem;
		border-bottom: 2px solid transparent;
	}
	nav a:hover,
	nav a.active {
		border-bottom-color: var(--accent);
	}

	@media (max-width: 640px) {
		header.site {
			padding: 1rem 1.25rem;
		}
		.logo {
			font-size: 1.35rem;
		}
		nav a {
			margin-left: 1rem;
			font-size: 0.8rem;
		}
	}
</style>
