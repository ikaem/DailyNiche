<!--
	Pill-style date navigation above the issue: prev/next day and a
	date-jump input. Purely static for now - no props, no wiring. Task
	6.1 gives it real prev/next/date-jump behavior once the API client
	exists to load a different day's posts.
-->
<div class="date-nav">
	<div class="pill">
		<!-- button, not a link - not yet wired to real navigation (Task 6.1) -->
		<button type="button"
			><span class="full">&larr; Previous</span><span class="short">&lsaquo;</span></button
		>
		<span class="current-date">
			<span class="full">Friday, July 10, 2026</span>
			<span class="short">Jul 10, 2026</span>
		</span>
		<input type="date" value="2026-07-10" />
		<!-- button, not a link - not yet wired to real navigation (Task 6.1) -->
		<button type="button"
			><span class="full">Next &rarr;</span><span class="short">&rsaquo;</span></button
		>
	</div>
</div>

<style>
	.date-nav {
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 1rem 2rem;
		gap: 0.75rem;
	}
	.date-nav .pill {
		display: inline-flex;
		align-items: center;
		gap: 0.75rem;
		background: var(--card);
		border: 1px solid var(--line);
		border-radius: 999px;
		padding: 0.5rem 0.5rem 0.5rem 1.1rem;
		box-shadow: 0 1px 3px rgba(24, 20, 15, 0.05);
	}
	.date-nav button {
		background: none;
		border: none;
		font-family: inherit;
		cursor: pointer;
		color: var(--ink-soft);
		font-size: 0.85rem;
		font-weight: 600;
		white-space: nowrap;
		padding: 0;
	}
	.date-nav button:hover {
		color: var(--accent);
	}
	.date-nav .current-date {
		font-weight: 600;
		font-size: 0.9rem;
		white-space: nowrap;
		padding: 0 0.25rem;
	}
	.date-nav input {
		border: none;
		background: var(--accent-soft);
		border-radius: 999px;
		padding: 0.3rem 0.5rem;
		color: var(--accent);
		font-size: 0.8rem;
	}
	.date-nav .short {
		display: none;
	}
	.date-nav .full {
		display: inline;
	}

	@media (max-width: 640px) {
		.date-nav {
			padding: 0.75rem 1rem;
		}
		.date-nav .pill {
			padding: 0.4rem 0.4rem 0.4rem 0.8rem;
			gap: 0.5rem;
		}
		.date-nav .full {
			display: none;
		}
		.date-nav .short {
			display: inline;
		}
		.date-nav .current-date {
			font-size: 0.8rem;
		}
		.date-nav input {
			width: 1.6rem;
			overflow: hidden;
		}
	}
</style>
