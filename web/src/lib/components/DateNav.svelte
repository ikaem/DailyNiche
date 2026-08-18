<!--
	Pill-style date navigation above the issue: prev/next day and a
	date-jump input. Navigates by updating the ?date= query param via
	goto(), which re-runs +page.server.ts's load - a plain server-rendered
	navigation, not a client-side fetch, consistent with how the rest of
	this app avoids the browser talking to data sources directly.
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { addDaysUTC, todayUTC } from '$lib/dateUtils';

	let { currentDate }: { currentDate: string } = $props();

	// There's no future data to show past today, so "Next" is disabled once
	// currentDate reaches it - todayUTC() is shared with +page.server.ts's own
	// copy via dateUtils.ts so client and server agree on what "today" means.
	const isToday = $derived(currentDate === todayUTC());

	function navigateTo(date: string) {
		goto(`?date=${date}`);
	}

	// Navigates with no ?date= param at all, rather than ?date=<today>, so
	// this reuses +page.server.ts's existing default path (getPostsToday())
	// instead of exercising the ?date= / getPostsByDate() path for what
	// should be the exact same data.
	function goToToday() {
		goto(page.url.pathname);
	}

	function goToPrevious() {
		navigateTo(addDaysUTC(currentDate, -1));
	}

	function goToNext() {
		navigateTo(addDaysUTC(currentDate, 1));
	}

	function onDateInputChange(event: Event) {
		const value = (event.target as HTMLInputElement).value;
		if (value) {
			navigateTo(value);
		}
	}

	// hr-HR to match postModel.ts's existing (if deferred-to-be-configurable,
	// per its own TODO) locale convention for dates elsewhere in the app -
	// the "Friday, July 10, 2026" this replaced was just placeholder mockup
	// text, not a deliberate choice to use English specifically here.
	const fullLabel = $derived(
		new Intl.DateTimeFormat('hr-HR', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		}).format(new Date(`${currentDate}T00:00:00Z`))
	);

	const shortLabel = $derived(
		new Intl.DateTimeFormat('hr-HR', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		}).format(new Date(`${currentDate}T00:00:00Z`))
	);
</script>

<div class="date-nav">
	<div class="pill">
		<button type="button" onclick={goToPrevious}
			><span class="full">&larr; Previous</span><span class="short">&lsaquo;</span></button
		>
		<span class="current-date">
			<span class="full">{fullLabel}</span>
			<span class="short">{shortLabel}</span>
		</span>
		<button type="button" class="today" onclick={goToToday}>Today</button>
		<input type="date" value={currentDate} onchange={onDateInputChange} />
		<button type="button" onclick={goToNext} disabled={isToday}
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
	.date-nav button:disabled {
		cursor: default;
		opacity: 0.4;
	}
	.date-nav button:disabled:hover {
		color: var(--ink-soft);
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
