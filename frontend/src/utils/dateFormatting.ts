/**
 * Date formatting utilities for session timestamps
 */

interface TimestampFormatOptions {
  includeYear?: boolean;
  use24Hour?: boolean;
}

/**
 * Formats a session timestamp into a human-readable format
 * 
 * @param timestamp - ISO timestamp string (e.g., "2024-09-03T20:20:00Z")
 * @param options - Optional formatting configuration
 * @returns Formatted string (e.g., "3 Sep, 08:20 pm" or "3 Sep 2023, 08:20 pm")
 */
export function formatSessionTimestamp(
  timestamp: string,
  options: TimestampFormatOptions = {}
): string {
  try {
    // Parse the timestamp
    const date = new Date(timestamp);
    
    // Check if the date is valid
    if (isNaN(date.getTime())) {
      console.warn(`Invalid timestamp provided: ${timestamp}`);
      return 'Invalid Date';
    }

    const currentYear = new Date().getFullYear();
    const timestampYear = date.getFullYear();
    
    // Determine if we should include the year
    const shouldIncludeYear = options.includeYear !== undefined 
      ? options.includeYear 
      : timestampYear !== currentYear;

    // Format the date part
    const monthNames = [
      'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
      'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
    ];
    
    const day = date.getDate();
    const month = monthNames[date.getMonth()];
    const year = timestampYear;

    // Format the time part
    const use24Hour = options.use24Hour || false;
    let timeString: string;
    
    if (use24Hour) {
      const hours = date.getHours().toString().padStart(2, '0');
      const minutes = date.getMinutes().toString().padStart(2, '0');
      timeString = `${hours}:${minutes}`;
    } else {
      // 12-hour format with am/pm
      let hours = date.getHours();
      const minutes = date.getMinutes().toString().padStart(2, '0');
      const ampm = hours >= 12 ? 'pm' : 'am';
      
      // Convert to 12-hour format
      hours = hours % 12;
      hours = hours ? hours : 12; // 0 should be 12
      
      timeString = `${hours}:${minutes} ${ampm}`;
    }

    // Combine date and time
    if (shouldIncludeYear) {
      return `${day} ${month} ${year}, ${timeString}`;
    } else {
      return `${day} ${month}, ${timeString}`;
    }
    
  } catch (error) {
    console.error(`Error formatting timestamp ${timestamp}:`, error);
    
    // Fallback to basic formatting
    try {
      const fallbackDate = new Date(timestamp);
      if (!isNaN(fallbackDate.getTime())) {
        return fallbackDate.toLocaleDateString() + ', ' + fallbackDate.toLocaleTimeString();
      }
    } catch (fallbackError) {
      console.error('Fallback formatting also failed:', fallbackError);
    }
    
    return 'Invalid Date';
  }
}

/**
 * Utility function to check if a timestamp is from the current year
 * 
 * @param timestamp - ISO timestamp string
 * @returns boolean indicating if the timestamp is from the current year
 */
export function isCurrentYear(timestamp: string): boolean {
  try {
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) {
      return false;
    }
    return date.getFullYear() === new Date().getFullYear();
  } catch (error) {
    console.error(`Error checking year for timestamp ${timestamp}:`, error);
    return false;
  }
}

/**
 * Utility function to validate if a timestamp string is valid
 * 
 * @param timestamp - ISO timestamp string to validate
 * @returns boolean indicating if the timestamp is valid
 */
export function isValidTimestamp(timestamp: string): boolean {
  try {
    const date = new Date(timestamp);
    return !isNaN(date.getTime());
  } catch (error) {
    return false;
  }
}