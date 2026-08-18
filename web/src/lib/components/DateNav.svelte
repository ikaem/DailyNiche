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
			><span class="arrow">&lsaquo;</span><span class="full">Previous</span></button
		>
		<span class="current-date">
			<span class="full">{fullLabel}</span>
			<span class="short">{shortLabel}</span>
		</span>
		<button type="button" class="today" onclick={goToToday}>Today</button>
		<span class="date-picker">
			<input
				type="date"
				value={currentDate}
				onchange={onDateInputChange}
				aria-label="Jump to date"
			/>
			<svg
				class="calendar-icon"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<rect x="3" y="5" width="18" height="16" rx="2" />
				<line x1="3" y1="10" x2="21" y2="10" />
				<line x1="8" y1="3" x2="8" y2="7" />
				<line x1="16" y1="3" x2="16" y2="7" />
			</svg>
		</span>
		<button type="button" onclick={goToNext} disabled={isToday}
			><span class="full">Next</span><span class="arrow">&rsaquo;</span></button
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
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
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
	.date-picker {
		position: relative;
		display: inline-flex;
	}
	/* Hidden on desktop, where the native input already shows its own
	   working icon next to the visible date text - only needed once the
	   mobile view collapses the input down to icon-only, below. */
	.date-picker .calendar-icon {
		display: none;
	}
	/* Previous/Next's arrow is a single always-visible element rather than
	   separate full/short copies (it looks identical either way) - only the
	   "Previous"/"Next" word next to it hides on small screens, via .full
	   above. The button's own flex layout (see .date-nav button) positions
	   it, rather than vertical-align, which turned out font/browser-
	   dependent (confirmed by comparing real screenshots across browsers). */

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
		/* The native picker-indicator icon's position inside a shrunk input is
		   browser/OS-controlled and not reliably centerable via padding
		   (confirmed empirically - it lands in a different spot in different
		   browsers, the same portability problem the arrow alignment fix hit
		   earlier). Instead: make the real input an invisible, full-circle
		   click target (still fully functional - clicking/tapping anywhere
		   on the circle opens the native date picker), and draw our own SVG
		   calendar icon on top, centered via ordinary flexbox - guaranteed
		   pixel-perfect in every browser since it's not relying on any
		   browser's internal form-control layout at all. */
		.date-picker {
			width: 1.6rem;
			height: 1.6rem;
		}
		.date-nav input {
			position: absolute;
			inset: 0;
			width: 100%;
			height: 100%;
			padding: 0;
			opacity: 0;
			cursor: pointer;
		}
		.date-picker .calendar-icon {
			display: flex;
			position: absolute;
			inset: 0;
			align-items: center;
			justify-content: center;
			width: 100%;
			height: 100%;
			padding: 0.4rem;
			border-radius: 999px;
			background: var(--accent-soft);
			color: var(--accent);
			pointer-events: none;
		}
	}
</style>
