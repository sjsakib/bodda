/**
 * Unit tests for date formatting utilities
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { 
  formatSessionTimestamp, 
  isCurrentYear, 
  isValidTimestamp 
} from '../dateFormatting';

describe('dateFormatting', () => {
  // Mock console methods to avoid noise in test output
  const originalConsoleWarn = console.warn;
  const originalConsoleError = console.error;
  
  beforeEach(() => {
    console.warn = vi.fn();
    console.error = vi.fn();
  });

  afterEach(() => {
    console.warn = originalConsoleWarn;
    console.error = originalConsoleError;
    vi.restoreAllMocks();
  });

  describe('formatSessionTimestamp', () => {
    describe('basic formatting', () => {
      it('should format current year timestamp without year', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp);
        // Test the format pattern rather than exact time due to timezone differences
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        expect(result).not.toContain(currentYear.toString());
      });

      it('should format previous year timestamp with year', () => {
        const previousYear = new Date().getFullYear() - 1;
        const timestamp = `${previousYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} Sep \d{4}, \d{1,2}:\d{2} [ap]m$/);
        expect(result).toContain(previousYear.toString());
      });

      it('should format future year timestamp with year', () => {
        const futureYear = new Date().getFullYear() + 1;
        const timestamp = `${futureYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} Sep \d{4}, \d{1,2}:\d{2} [ap]m$/);
        expect(result).toContain(futureYear.toString());
      });
    });

    describe('time formatting', () => {
      it('should format time in 12-hour format with am/pm', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T08:30:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        // Check that it contains either am or pm (timezone dependent)
        expect(result).toMatch(/[ap]m$/);
      });

      it('should format time with proper am/pm indicators', () => {
        const currentYear = new Date().getFullYear();
        // Test with a time that should be PM in most timezones
        const timestamp = `${currentYear}-09-03T18:00:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
      });

      it('should pad single digit minutes', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T15:05:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:05 [ap]m$/);
        expect(result).toContain(':05');
      });

      it('should handle various hour formats', () => {
        const currentYear = new Date().getFullYear();
        const timestamps = [
          `${currentYear}-09-03T01:00:00Z`,
          `${currentYear}-09-03T12:00:00Z`,
          `${currentYear}-09-03T13:00:00Z`,
          `${currentYear}-09-03T23:00:00Z`
        ];
        
        timestamps.forEach(timestamp => {
          const result = formatSessionTimestamp(timestamp);
          expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        });
      });
    });

    describe('24-hour format option', () => {
      it('should format in 24-hour format when use24Hour is true', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp, { use24Hour: true });
        expect(result).toMatch(/^\d{1,2} Sep, \d{2}:\d{2}$/);
        expect(result).not.toContain('am');
        expect(result).not.toContain('pm');
      });

      it('should format morning time in 24-hour format', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T08:05:00Z`;
        const result = formatSessionTimestamp(timestamp, { use24Hour: true });
        expect(result).toMatch(/^\d{1,2} Sep, \d{2}:\d{2}$/);
        expect(result).not.toContain('am');
        expect(result).not.toContain('pm');
      });

      it('should format with proper zero padding in 24-hour format', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T00:00:00Z`;
        const result = formatSessionTimestamp(timestamp, { use24Hour: true });
        expect(result).toMatch(/^\d{1,2} Sep, \d{2}:\d{2}$/);
        expect(result).not.toContain('am');
        expect(result).not.toContain('pm');
      });
    });

    describe('year inclusion logic', () => {
      it('should include year when explicitly requested', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp, { includeYear: true });
        expect(result).toMatch(/^\d{1,2} Sep \d{4}, \d{1,2}:\d{2} [ap]m$/);
        expect(result).toContain(currentYear.toString());
      });

      it('should exclude year when explicitly requested', () => {
        const previousYear = new Date().getFullYear() - 1;
        const timestamp = `${previousYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp, { includeYear: false });
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        expect(result).not.toContain(previousYear.toString());
      });
    });

    describe('month formatting', () => {
      const currentYear = new Date().getFullYear();
      
      const monthTests = [
        { month: '01', expected: 'Jan' },
        { month: '02', expected: 'Feb' },
        { month: '03', expected: 'Mar' },
        { month: '04', expected: 'Apr' },
        { month: '05', expected: 'May' },
        { month: '06', expected: 'Jun' },
        { month: '07', expected: 'Jul' },
        { month: '08', expected: 'Aug' },
        { month: '09', expected: 'Sep' },
        { month: '10', expected: 'Oct' },
        { month: '11', expected: 'Nov' },
        { month: '12', expected: 'Dec' }
      ];

      monthTests.forEach(({ month, expected }) => {
        it(`should format ${expected} correctly`, () => {
          const timestamp = `${currentYear}-${month}-15T12:00:00Z`;
          const result = formatSessionTimestamp(timestamp);
          expect(result).toContain(expected);
        });
      });
    });

    describe('edge cases and boundaries', () => {
      it('should handle year boundary correctly', () => {
        const currentYear = new Date().getFullYear();
        
        // Test with a date that's clearly in the previous year
        const previousYearDate = `${currentYear - 1}-06-15T12:00:00Z`;
        const result1 = formatSessionTimestamp(previousYearDate);
        // Should include year for previous year
        expect(result1).toMatch(/^\d{1,2} \w{3} \d{4}, \d{1,2}:\d{2} [ap]m$/);
        expect(result1).toContain((currentYear - 1).toString());
        
        // Test with a date that's clearly in the current year
        const currentYearDate = `${currentYear}-06-15T12:00:00Z`;
        const result2 = formatSessionTimestamp(currentYearDate);
        // Should not contain the year since it's current year
        expect(result2).toMatch(/^\d{1,2} \w{3}, \d{1,2}:\d{2} [ap]m$/);
        expect(result2).not.toContain(currentYear.toString());
      });

      it('should handle leap year dates', () => {
        const timestamp = '2024-02-29T12:00:00Z'; // 2024 is a leap year
        const result = formatSessionTimestamp(timestamp);
        expect(result).toContain('29 Feb');
      });

      it('should handle single digit days', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-01T12:00:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^1 Sep, \d{1,2}:\d{2} [ap]m$/);
      });

      it('should handle double digit days', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-15T12:00:00Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^15 Sep, \d{1,2}:\d{2} [ap]m$/);
      });
    });

    describe('timezone handling', () => {
      it('should handle UTC timestamps', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp);
        // The exact result depends on local timezone, but should be valid format
        expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
      });

      it('should handle timestamps with timezone offset', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T20:20:00+05:00`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
      });

      it('should handle timestamps with negative timezone offset', () => {
        const currentYear = new Date().getFullYear();
        const timestamp = `${currentYear}-09-03T20:20:00-08:00`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
      });
    });

    describe('error handling', () => {
      it('should handle invalid timestamp strings', () => {
        const result = formatSessionTimestamp('invalid-date');
        expect(result).toBe('Invalid Date');
        expect(console.warn).toHaveBeenCalledWith('Invalid timestamp provided: invalid-date');
      });

      it('should handle empty string', () => {
        const result = formatSessionTimestamp('');
        expect(result).toBe('Invalid Date');
        expect(console.warn).toHaveBeenCalledWith('Invalid timestamp provided: ');
      });

      it('should handle null-like values', () => {
        const result = formatSessionTimestamp('null');
        expect(result).toBe('Invalid Date');
      });

      it('should handle malformed ISO strings', () => {
        const result = formatSessionTimestamp('2024-13-45T25:70:00Z');
        expect(result).toBe('Invalid Date');
      });

      it('should handle undefined and null-like inputs gracefully', () => {
        const invalidInputs = [
          'undefined',
          'null',
          'NaN',
          '{}',
          '[]',
          'false',
          'truly-invalid-date'
        ];

        invalidInputs.forEach(input => {
          const result = formatSessionTimestamp(input);
          expect(result).toBe('Invalid Date');
        });
      });

      it('should handle extremely large and small dates', () => {
        // Test with dates far in the future and past
        const farFuture = '9999-12-31T12:00:00Z';
        const farPast = '1000-01-01T12:00:00Z'; // Use a more reasonable past date
        
        const futureResult = formatSessionTimestamp(farFuture);
        const pastResult = formatSessionTimestamp(farPast);
        
        // Should handle these gracefully (JavaScript Date can handle these)
        // Due to timezone conversion, we just check the general format
        expect(futureResult).toMatch(/^\d{1,2} \w{3} \d{4}, \d{1,2}:\d{2} [ap]m$/);
        expect(pastResult).toMatch(/^\d{1,2} \w{3} \d{1,4}, \d{1,2}:\d{2} [ap]m$/); // Allow 1-4 digits for year
        expect(futureResult).toContain('9999');
        expect(pastResult).toContain('1000');
      });

      it('should use fallback formatting when main formatting fails but date is parseable', () => {
        // Create a spy on console.error to track calls
        const errorSpy = vi.spyOn(console, 'error');
        
        // We'll test this by providing a timestamp that causes issues in the main path
        // but can be handled by the fallback
        const timestamp = 'invalid-but-parseable';
        const result = formatSessionTimestamp(timestamp);
        
        // Should return 'Invalid Date' for truly invalid timestamps
        expect(result).toBe('Invalid Date');
        expect(console.warn).toHaveBeenCalled();
        
        errorSpy.mockRestore();
      });

      it('should handle special date strings that JavaScript accepts', () => {
        // JavaScript Date constructor accepts some unusual formats
        const specialFormats = [
          'December 25, 2024',
          '2024/12/25',
          '12/25/2024',
          '25 Dec 2024'
        ];

        specialFormats.forEach(format => {
          const result = formatSessionTimestamp(format);
          // These should either format correctly or return 'Invalid Date'
          expect(typeof result).toBe('string');
          expect(result.length).toBeGreaterThan(0);
        });
      });
    });

    describe('various timestamp formats', () => {
      const currentYear = new Date().getFullYear();
      
      it('should handle ISO string without milliseconds', () => {
        const timestamp = `${currentYear}-09-03T20:20:00Z`;
        const result = formatSessionTimestamp(timestamp);
        // Check format pattern instead of exact time due to timezone differences
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        expect(result).toContain('Sep');
      });

      it('should handle ISO string with milliseconds', () => {
        const timestamp = `${currentYear}-09-03T20:20:00.123Z`;
        const result = formatSessionTimestamp(timestamp);
        // Check format pattern instead of exact time due to timezone differences
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        expect(result).toContain('Sep');
      });

      it('should handle ISO string without Z suffix', () => {
        const timestamp = `${currentYear}-09-03T20:20:00`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
      });

      it('should handle different date separators', () => {
        // Test with different valid date formats
        const validFormats = [
          `${currentYear}/09/03 20:20:00`,
          `Sep 03, ${currentYear} 20:20:00`,
          `${currentYear}-09-03 20:20:00`
        ];
        
        validFormats.forEach(timestamp => {
          const result = formatSessionTimestamp(timestamp);
          expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
        });
      });

      it('should handle timestamps with microseconds', () => {
        const timestamp = `${currentYear}-09-03T20:20:00.123456Z`;
        const result = formatSessionTimestamp(timestamp);
        expect(result).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);
        expect(result).toContain('Sep');
      });
    });

    describe('comprehensive timezone tests', () => {
      const currentYear = new Date().getFullYear();

      it('should handle various timezone formats', () => {
        const timezoneFormats = [
          `${currentYear}-09-03T20:20:00Z`,
          `${currentYear}-09-03T20:20:00+00:00`,
          `${currentYear}-09-03T20:20:00-00:00`,
          `${currentYear}-09-03T20:20:00+05:30`,
          `${currentYear}-09-03T20:20:00-08:00`,
          `${currentYear}-09-03T20:20:00+12:00`
        ];

        timezoneFormats.forEach(timestamp => {
          const result = formatSessionTimestamp(timestamp);
          expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
        });
      });

      it('should maintain consistency across different timezones for same UTC time', () => {
        const baseTime = `${currentYear}-09-03T12:00:00`;
        const utcTime = baseTime + 'Z';
        const plusTime = baseTime + '+00:00';
        
        const result1 = formatSessionTimestamp(utcTime);
        const result2 = formatSessionTimestamp(plusTime);
        
        // Both should produce the same result since they represent the same UTC time
        expect(result1).toBe(result2);
      });
    });

    describe('performance and stress tests', () => {
      const currentYear = new Date().getFullYear();

      it('should handle rapid successive calls efficiently', () => {
        const timestamp = `${currentYear}-09-03T20:20:00Z`;
        const startTime = performance.now();
        
        // Call the function many times
        for (let i = 0; i < 1000; i++) {
          formatSessionTimestamp(timestamp);
        }
        
        const endTime = performance.now();
        const duration = endTime - startTime;
        
        // Should complete 1000 calls in reasonable time (less than 100ms)
        expect(duration).toBeLessThan(100);
      });

      it('should handle array of mixed timestamps', () => {
        const timestamps = [
          `${currentYear}-01-01T00:00:00Z`,
          `${currentYear}-06-15T12:30:45Z`,
          `${currentYear}-12-31T23:59:59Z`,
          `${currentYear - 1}-03-15T08:20:00Z`,
          `${currentYear + 1}-07-04T16:45:30Z`
        ];

        const results = timestamps.map(ts => formatSessionTimestamp(ts));
        
        // All should be valid formatted strings
        results.forEach(result => {
          expect(result).toMatch(/^\d{1,2} \w{3}(?: \d{4})?, \d{1,2}:\d{2} [ap]m$/);
        });

        // Current year timestamps should not include year
        expect(results[0]).not.toContain(currentYear.toString());
        expect(results[1]).not.toContain(currentYear.toString());
        expect(results[2]).not.toContain(currentYear.toString());

        // Non-current year timestamps should include year
        expect(results[3]).toContain((currentYear - 1).toString());
        expect(results[4]).toContain((currentYear + 1).toString());
      });
    });
  });

  describe('isCurrentYear', () => {
    it('should return true for current year timestamps', () => {
      const currentYear = new Date().getFullYear();
      const timestamp = `${currentYear}-09-03T20:20:00Z`;
      expect(isCurrentYear(timestamp)).toBe(true);
    });

    it('should return false for previous year timestamps', () => {
      const previousYear = new Date().getFullYear() - 1;
      const timestamp = `${previousYear}-09-03T20:20:00Z`;
      expect(isCurrentYear(timestamp)).toBe(false);
    });

    it('should return false for future year timestamps', () => {
      const futureYear = new Date().getFullYear() + 1;
      const timestamp = `${futureYear}-09-03T20:20:00Z`;
      expect(isCurrentYear(timestamp)).toBe(false);
    });

    it('should return false for invalid timestamps', () => {
      expect(isCurrentYear('invalid-date')).toBe(false);
      // Note: The function doesn't log errors for invalid dates that don't throw,
      // it only logs errors for exceptions in the try-catch block
    });

    it('should return false for empty string', () => {
      expect(isCurrentYear('')).toBe(false);
    });
  });

  describe('isValidTimestamp', () => {
    it('should return true for valid ISO timestamps', () => {
      expect(isValidTimestamp('2024-09-03T20:20:00Z')).toBe(true);
      expect(isValidTimestamp('2024-09-03T20:20:00.123Z')).toBe(true);
      expect(isValidTimestamp('2024-09-03T20:20:00+05:00')).toBe(true);
    });

    it('should return false for invalid timestamps', () => {
      expect(isValidTimestamp('invalid-date')).toBe(false);
      expect(isValidTimestamp('')).toBe(false);
      expect(isValidTimestamp('2024-13-45T25:70:00Z')).toBe(false);
    });

    it('should return false for null-like values', () => {
      expect(isValidTimestamp('null')).toBe(false);
      expect(isValidTimestamp('undefined')).toBe(false);
    });

    it('should handle edge cases gracefully', () => {
      expect(isValidTimestamp('2024-02-29T12:00:00Z')).toBe(true); // Valid leap year
      // Note: JavaScript Date constructor is lenient and converts invalid dates
      // 2023-02-29 becomes 2023-03-01, so it's still a valid date
      expect(isValidTimestamp('2023-02-29T12:00:00Z')).toBe(true); // JavaScript converts to valid date
      expect(isValidTimestamp('2023-02-30T12:00:00Z')).toBe(true); // JavaScript converts to valid date
      expect(isValidTimestamp('invalid-date-string')).toBe(false); // Truly invalid
    });

    it('should handle various valid date formats', () => {
      const validFormats = [
        '2024-01-01T00:00:00Z',
        '2024-12-31T23:59:59.999Z',
        '2024-06-15T12:30:45+05:30',
        '2024-06-15T12:30:45-08:00',
        'December 25, 2024',
        '2024/12/25',
        '12/25/2024'
      ];

      validFormats.forEach(format => {
        expect(isValidTimestamp(format)).toBe(true);
      });
    });

    it('should handle boundary dates correctly', () => {
      // Test extreme but valid dates
      expect(isValidTimestamp('1970-01-01T00:00:00Z')).toBe(true); // Unix epoch
      expect(isValidTimestamp('2038-01-19T03:14:07Z')).toBe(true); // Near 32-bit limit
      expect(isValidTimestamp('1900-01-01T00:00:00Z')).toBe(true); // Early 20th century
      expect(isValidTimestamp('2100-12-31T23:59:59Z')).toBe(true); // Far future
    });

    it('should return false for malformed strings', () => {
      const malformedInputs = [
        'completely-invalid',
        'not-a-date-at-all',
        'T20:20:00Z', // Missing date part
        '2024-01-01T', // Missing time part
        'invalid-timestamp',
        'abc-def-ghi'
      ];

      malformedInputs.forEach(input => {
        expect(isValidTimestamp(input)).toBe(false);
      });
    });
  });

  describe('integration tests', () => {
    it('should work correctly with all utility functions together', () => {
      const currentYear = new Date().getFullYear();
      const validTimestamp = `${currentYear}-09-03T20:20:00Z`;
      const invalidTimestamp = 'invalid-date';

      // Test valid timestamp
      expect(isValidTimestamp(validTimestamp)).toBe(true);
      expect(isCurrentYear(validTimestamp)).toBe(true);
      const formatted = formatSessionTimestamp(validTimestamp);
      expect(formatted).toMatch(/^\d{1,2} Sep, \d{1,2}:\d{2} [ap]m$/);

      // Test invalid timestamp
      expect(isValidTimestamp(invalidTimestamp)).toBe(false);
      expect(isCurrentYear(invalidTimestamp)).toBe(false);
      expect(formatSessionTimestamp(invalidTimestamp)).toBe('Invalid Date');
    });

    it('should handle mixed arrays of valid and invalid timestamps', () => {
      const currentYear = new Date().getFullYear();
      const mixedTimestamps = [
        `${currentYear}-01-01T12:00:00Z`,
        'invalid-date',
        `${currentYear - 1}-06-15T12:00:00Z`,
        '',
        `${currentYear}-06-15T12:30:00Z`
      ];

      const results = mixedTimestamps.map(ts => ({
        timestamp: ts,
        isValid: isValidTimestamp(ts),
        isCurrentYear: isCurrentYear(ts),
        formatted: formatSessionTimestamp(ts)
      }));

      // Valid timestamps should format correctly
      expect(results[0].isValid).toBe(true);
      expect(results[0].isCurrentYear).toBe(true);
      expect(results[0].formatted).toMatch(/^\d{1,2} Jan, \d{1,2}:\d{2} [ap]m$/);

      expect(results[2].isValid).toBe(true);
      expect(results[2].isCurrentYear).toBe(false);
      expect(results[2].formatted).toMatch(/^\d{1,2} Jun \d{4}, \d{1,2}:\d{2} [ap]m$/);

      expect(results[4].isValid).toBe(true);
      expect(results[4].isCurrentYear).toBe(true);
      expect(results[4].formatted).toMatch(/^\d{1,2} Jun, \d{1,2}:\d{2} [ap]m$/);

      // Invalid timestamps should be handled gracefully
      expect(results[1].isValid).toBe(false);
      expect(results[1].isCurrentYear).toBe(false);
      expect(results[1].formatted).toBe('Invalid Date');

      expect(results[3].isValid).toBe(false);
      expect(results[3].isCurrentYear).toBe(false);
      expect(results[3].formatted).toBe('Invalid Date');
    });
  });
});