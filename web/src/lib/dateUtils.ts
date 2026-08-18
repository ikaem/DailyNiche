// Adds `days` (positive or negative) to a YYYY-MM-DD date string, returning
// a new YYYY-MM-DD string - computed in UTC to stay consistent with this
// app's UTC-everywhere convention (see CLAUDE.md's Timestamps & Timezones
// note), so date-boundary arithmetic doesn't drift with the browser's local
// timezone.
export function addDaysUTC(dateStr: string, days: number): string {
	const date = new Date(`${dateStr}T00:00:00Z`);
	date.setUTCDate(date.getUTCDate() + days);
	return date.toISOString().slice(0, 10);
}

// Today's date in UTC as YYYY-MM-DD - shared here so both server (+page.server.ts)
// and client (DateNav.svelte, for disabling "Next" once already on today) agree
// on what "today" means, rather than each computing it independently.
export function todayUTC(): string {
	return new Date().toISOString().slice(0, 10);
}
