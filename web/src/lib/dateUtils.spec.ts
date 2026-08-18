import { afterEach, describe, expect, it, vi } from 'vitest';
import { addDaysUTC, todayUTC } from './dateUtils';

describe('addDaysUTC', () => {
	it('adds a positive number of days within the same month', () => {
		// given: a date early in the month
		// when: adding 3 days
		const result = addDaysUTC('2026-07-10', 3);

		// then: it lands 3 days later
		expect(result).toBe('2026-07-13');
	});

	it('subtracts days via a negative number', () => {
		// given: a date
		// when: adding -1 day
		const result = addDaysUTC('2026-07-10', -1);

		// then: it lands on the previous day
		expect(result).toBe('2026-07-09');
	});

	it('rolls over into the next month', () => {
		// given: the last day of July
		// when: adding 1 day
		const result = addDaysUTC('2026-07-31', 1);

		// then: it rolls into August
		expect(result).toBe('2026-08-01');
	});

	it('rolls over into the previous month', () => {
		// given: the first day of a month
		// when: subtracting 1 day
		const result = addDaysUTC('2026-08-01', -1);

		// then: it rolls back into July
		expect(result).toBe('2026-07-31');
	});

	it('rolls over across a year boundary', () => {
		// given: the last day of the year
		// when: adding 1 day
		const result = addDaysUTC('2026-12-31', 1);

		// then: it rolls into the next year
		expect(result).toBe('2027-01-01');
	});
});

describe('todayUTC', () => {
	afterEach(() => {
		vi.useRealTimers();
	});

	it('returns the current date in UTC as YYYY-MM-DD', () => {
		// given: a frozen system time
		vi.setSystemTime(new Date('2026-07-15T23:30:00Z'));

		// when: asking for today's date
		const result = todayUTC();

		// then: it reflects the UTC date, not shifted by any local timezone
		expect(result).toBe('2026-07-15');
	});
});
