import { page } from 'vitest/browser';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import FeedDaysFilter from './FeedDaysFilter.svelte';

describe('FeedDaysFilter.svelte', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	afterEach(() => {
		localStorage.clear();
	});

	it('renders unchecked by default with the days input disabled', async () => {
		// given/when: rendered with no props and no persisted state
		render(FeedDaysFilter);

		// then: off by default, days input reflects the default and is disabled
		const checkbox = page.getByRole('checkbox', { name: 'Limit feed posts to' });
		await expect.element(checkbox).not.toBeChecked();

		const daysInput = page.getByLabelText('Number of days');
		await expect.element(daysInput).toBeDisabled();
		await expect.element(daysInput).toHaveValue(1);
	});

	it('enables the days input and persists the toggle when checked', async () => {
		// given: the component rendered in its default (off) state
		render(FeedDaysFilter);
		const checkbox = page.getByRole('checkbox', { name: 'Limit feed posts to' });

		// when: the checkbox is checked
		await checkbox.click();

		// then: the days input becomes enabled, and the choice is persisted
		await expect.element(checkbox).toBeChecked();
		await expect.element(page.getByLabelText('Number of days')).toBeEnabled();
		expect(localStorage.getItem('feedDaysFilterEnabled')).toBe('true');
	});

	it('persists a changed day count', async () => {
		// given: the component rendered already enabled
		render(FeedDaysFilter, { enabled: true });
		const daysInput = page.getByLabelText('Number of days');

		// when: the day count is changed
		await daysInput.fill('7');

		// then: the new value is persisted
		await expect.element(daysInput).toHaveValue(7);
		expect(localStorage.getItem('feedDaysFilterDays')).toBe('7');
	});

	it('restores a previously persisted enabled state and day count on mount', async () => {
		// given: a browser that already has a preference saved from before
		localStorage.setItem('feedDaysFilterEnabled', 'true');
		localStorage.setItem('feedDaysFilterDays', '10');

		// when: the component mounts fresh
		render(FeedDaysFilter);

		// then: it picks up the persisted values, not its own defaults
		const checkbox = page.getByRole('checkbox', { name: 'Limit feed posts to' });
		await expect.element(checkbox).toBeChecked();
		await expect.element(page.getByLabelText('Number of days')).toHaveValue(10);
	});
});
