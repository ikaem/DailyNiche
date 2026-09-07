<!--
	Lets the reader cap how many days back a feed's posts can be, to tame
	the backlog dump a newly-added feed's first fetch produces (every post
	gets fetchedAt "today", but publishedAt can span months back - see
	CLAUDE.md's design discussion for the full reasoning). Off by default,
	relative to whichever issue is being viewed (see postFilters.ts), not
	real-world today.

	Persisted to localStorage, not any server-side store - there are no
	user accounts in this app, so a per-browser display preference like
	this has nowhere else to live.
-->
<script lang="ts">
	const STORAGE_KEY_ENABLED = 'feedDaysFilterEnabled';
	const STORAGE_KEY_DAYS = 'feedDaysFilterDays';

	let { enabled = $bindable(false), days = $bindable(1) }: { enabled?: boolean; days?: number } =
		$props();

	// Runs once on mount, browser-only - localStorage doesn't exist during
	// SSR, and $effect itself never runs during SSR either, so this is a
	// safe place to restore whatever this browser last had set.
	$effect(() => {
		const storedEnabled = localStorage.getItem(STORAGE_KEY_ENABLED);
		if (storedEnabled !== null) enabled = storedEnabled === 'true';

		const storedDays = Number(localStorage.getItem(STORAGE_KEY_DAYS));
		if (Number.isFinite(storedDays) && storedDays > 0) days = storedDays;
	});

	$effect(() => {
		localStorage.setItem(STORAGE_KEY_ENABLED, String(enabled));
	});

	$effect(() => {
		localStorage.setItem(STORAGE_KEY_DAYS, String(days));
	});

	// Manual handler rather than bind:value - a raw two-way bind would
	// coerce an empty/invalid in-progress edit straight to NaN and write
	// that into `days` (and localStorage) immediately. Only accept a
	// change once it's a real positive number.
	function onDaysInput(event: Event) {
		const value = Number((event.target as HTMLInputElement).value);
		if (Number.isFinite(value) && value > 0) {
			days = value;
		}
	}
</script>

<div class="feed-days-filter">
	<label>
		<input type="checkbox" bind:checked={enabled} />
		Limit feed posts to
	</label>
	<input
		type="number"
		min="1"
		value={days}
		oninput={onDaysInput}
		disabled={!enabled}
		aria-label="Number of days"
	/>
	<span>days</span>
</div>

<style>
	.feed-days-filter {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.5rem;
		padding: 0 2rem 1rem;
		font-size: 0.85rem;
		color: var(--ink-soft);
	}
	.feed-days-filter label {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		cursor: pointer;
	}
	.feed-days-filter input[type='number'] {
		width: 3.5rem;
		border: 1px solid var(--line);
		border-radius: 6px;
		padding: 0.2rem 0.4rem;
		font-family: inherit;
		font-size: inherit;
		color: inherit;
	}
	.feed-days-filter input[type='number']:disabled {
		opacity: 0.5;
	}
</style>
