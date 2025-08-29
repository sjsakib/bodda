import React from 'react';

/**
 * Configuration for diagram rendering timeouts
 */
export interface DiagramTimeoutConfig {
  /** Timeout duration in milliseconds */
  timeout: number;
  /** Whether to enable timeout handling */
  enabled: boolean;
  /** Custom timeout message */
  message?: string;
}

/**
 * Default timeout configurations for different diagram types
 */
export const DEFAULT_TIMEOUT_CONFIG: Record<string, DiagramTimeoutConfig> = {
  mermaid: {
    timeout: 10000, // 10 seconds
    enabled: true,
    message: 'Mermaid diagram rendering timeout',
  },
  'vega-lite': {
    timeout: 15000, // 15 seconds (charts can be more complex)
    enabled: true,
    message: 'Vega-Lite chart rendering timeout',
  },
  library: {
    timeout: 30000, // 30 seconds for library loading
    enabled: true,
    message: 'Diagram library loading timeout',
  },
};

/**
 * Error class for timeout-related errors
 */
export class DiagramTimeoutError extends Error {
  public readonly isTimeout = true;
  public readonly diagramType: string;
  public readonly duration: number;

  constructor(diagramType: string, duration: number, message?: string) {
    super(message || `${diagramType} diagram rendering timeout after ${duration}ms`);
    this.name = 'DiagramTimeoutError';
    this.diagramType = diagramType;
    this.duration = duration;
  }
}

/**
 * Utility class for managing diagram rendering timeouts
 */
export class DiagramTimeoutManager {
  private timeouts = new Map<string, NodeJS.Timeout>();
  private startTimes = new Map<string, number>();

  /**
   * Start a timeout for a specific operation
   */
  startTimeout(
    operationId: string,
    config: DiagramTimeoutConfig,
    diagramType: string
  ): Promise<never> {
    if (!config.enabled) {
      return new Promise(() => {}); // Never resolves if timeout is disabled
    }

    // Clear any existing timeout for this operation
    this.clearTimeout(operationId);

    // Record start time
    this.startTimes.set(operationId, Date.now());

    return new Promise((_, reject) => {
      const timeoutId = setTimeout(() => {
        const duration = Date.now() - (this.startTimes.get(operationId) || 0);
        this.clearTimeout(operationId);
        
        const error = new DiagramTimeoutError(
          diagramType,
          duration,
          config.message
        );
        
        console.warn(`Diagram timeout: ${operationId}`, {
          diagramType,
          duration,
          configuredTimeout: config.timeout,
        });
        
        reject(error);
      }, config.timeout);

      this.timeouts.set(operationId, timeoutId);
    });
  }

  /**
   * Clear a specific timeout
   */
  clearTimeout(operationId: string): void {
    const timeoutId = this.timeouts.get(operationId);
    if (timeoutId) {
      clearTimeout(timeoutId);
      this.timeouts.delete(operationId);
      this.startTimes.delete(operationId);
    }
  }

  /**
   * Clear all timeouts
   */
  clearAllTimeouts(): void {
    this.timeouts.forEach((timeoutId) => clearTimeout(timeoutId));
    this.timeouts.clear();
    this.startTimes.clear();
  }

  /**
   * Get the elapsed time for an operation
   */
  getElapsedTime(operationId: string): number {
    const startTime = this.startTimes.get(operationId);
    return startTime ? Date.now() - startTime : 0;
  }

  /**
   * Check if an operation is currently timing out
   */
  hasActiveTimeout(operationId: string): boolean {
    return this.timeouts.has(operationId);
  }

  /**
   * Get all active timeout operation IDs
   */
  getActiveTimeouts(): string[] {
    return Array.from(this.timeouts.keys());
  }
}

/**
 * Global timeout manager instance
 */
export const globalTimeoutManager = new DiagramTimeoutManager();

/**
 * Utility function to wrap a promise with timeout handling
 */
export async function withTimeout<T>(
  promise: Promise<T>,
  operationId: string,
  diagramType: string,
  config?: Partial<DiagramTimeoutConfig>
): Promise<T> {
  const timeoutConfig = {
    ...DEFAULT_TIMEOUT_CONFIG[diagramType],
    ...config,
  };

  if (!timeoutConfig.enabled) {
    return promise;
  }

  const timeoutPromise = globalTimeoutManager.startTimeout(
    operationId,
    timeoutConfig,
    diagramType
  );

  try {
    const result = await Promise.race([promise, timeoutPromise]);
    globalTimeoutManager.clearTimeout(operationId);
    return result;
  } catch (error) {
    globalTimeoutManager.clearTimeout(operationId);
    throw error;
  }
}

/**
 * Utility function to create a timeout-aware operation wrapper
 */
export function createTimeoutWrapper<T extends any[], R>(
  fn: (...args: T) => Promise<R>,
  diagramType: string,
  config?: Partial<DiagramTimeoutConfig>
) {
  return async (...args: T): Promise<R> => {
    const operationId = `${diagramType}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    try {
      const promise = fn(...args);
      return await withTimeout(promise, operationId, diagramType, config);
    } catch (error) {
      // Add timeout context to non-timeout errors
      if (error instanceof DiagramTimeoutError) {
        throw error;
      }
      
      // Wrap other errors with timeout context if they occurred during timeout period
      if (globalTimeoutManager.hasActiveTimeout(operationId)) {
        const elapsedTime = globalTimeoutManager.getElapsedTime(operationId);
        console.warn(`Operation failed during timeout period: ${operationId}`, {
          diagramType,
          elapsedTime,
          originalError: error,
        });
      }
      
      throw error;
    }
  };
}

/**
 * Hook for managing timeouts in React components
 */
export function useTimeoutManager() {
  const [activeTimeouts, setActiveTimeouts] = React.useState<string[]>([]);

  React.useEffect(() => {
    const interval = setInterval(() => {
      setActiveTimeouts(globalTimeoutManager.getActiveTimeouts());
    }, 1000);

    return () => {
      clearInterval(interval);
      globalTimeoutManager.clearAllTimeouts();
    };
  }, []);

  const startTimeout = React.useCallback(
    (operationId: string, config: DiagramTimeoutConfig, diagramType: string) => {
      return globalTimeoutManager.startTimeout(operationId, config, diagramType);
    },
    []
  );

  const clearTimeout = React.useCallback((operationId: string) => {
    globalTimeoutManager.clearTimeout(operationId);
  }, []);

  const clearAllTimeouts = React.useCallback(() => {
    globalTimeoutManager.clearAllTimeouts();
  }, []);

  const getElapsedTime = React.useCallback((operationId: string) => {
    return globalTimeoutManager.getElapsedTime(operationId);
  }, []);

  return {
    activeTimeouts,
    startTimeout,
    clearTimeout,
    clearAllTimeouts,
    getElapsedTime,
  };
}

/**
 * Utility to check if an error is a timeout error
 */
export function isTimeoutError(error: any): error is DiagramTimeoutError {
  return error instanceof DiagramTimeoutError || 
         (error && error.isTimeout === true) ||
         (error && typeof error.message === 'string' && error.message.toLowerCase().includes('timeout'));
}

/**
 * Utility to get timeout configuration for a diagram type
 */
export function getTimeoutConfig(
  diagramType: string,
  overrides?: Partial<DiagramTimeoutConfig>
): DiagramTimeoutConfig {
  const defaultConfig = DEFAULT_TIMEOUT_CONFIG[diagramType] || DEFAULT_TIMEOUT_CONFIG.mermaid;
  return { ...defaultConfig, ...overrides };
}

/**
 * Utility to format timeout duration for display
 */
export function formatTimeoutDuration(milliseconds: number): string {
  if (milliseconds < 1000) {
    return `${milliseconds}ms`;
  }
  
  const seconds = Math.floor(milliseconds / 1000);
  const remainingMs = milliseconds % 1000;
  
  if (remainingMs === 0) {
    return `${seconds}s`;
  }
  
  return `${seconds}.${Math.floor(remainingMs / 100)}s`;
}

export default DiagramTimeoutManager;