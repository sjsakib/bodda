/**
 * Diagram caching utilities for performance optimization
 * Implements LRU cache with content hashing to avoid re-rendering identical diagrams
 */

/**
 * Interface for cached diagram entries
 */
interface CachedDiagram {
  /** Rendered diagram content (SVG string or spec object) */
  content: any;
  /** Timestamp when cached */
  timestamp: number;
  /** Access count for LRU eviction */
  accessCount: number;
  /** Last access timestamp */
  lastAccessed: number;
  /** Size estimate in bytes */
  size: number;
}

/**
 * Cache configuration options
 */
interface CacheConfig {
  /** Maximum number of cached diagrams */
  maxEntries: number;
  /** Maximum cache size in bytes */
  maxSize: number;
  /** TTL in milliseconds */
  ttl: number;
}

/**
 * Default cache configuration
 */
const DEFAULT_CONFIG: CacheConfig = {
  maxEntries: 100,
  maxSize: 10 * 1024 * 1024, // 10MB
  ttl: 30 * 60 * 1000, // 30 minutes
};

/**
 * Simple hash function for content
 */
const hashContent = (content: string): string => {
  let hash = 0;
  for (let i = 0; i < content.length; i++) {
    const char = content.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32-bit integer
  }
  return hash.toString(36);
};

/**
 * Estimate size of cached content in bytes
 */
const estimateSize = (content: any): number => {
  if (typeof content === 'string') {
    return content.length * 2; // UTF-16 encoding
  }
  return JSON.stringify(content).length * 2;
};

/**
 * LRU Cache implementation for diagram rendering results
 */
class DiagramCache {
  private cache = new Map<string, CachedDiagram>();
  private config: CacheConfig;
  private currentSize = 0;

  constructor(config: Partial<CacheConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /**
   * Generate cache key from content and options
   */
  private generateKey(content: string, type: string, theme?: string, options?: any): string {
    const contentHash = hashContent(content);
    const optionsHash = options ? hashContent(JSON.stringify(options)) : '';
    return `${type}:${theme || 'default'}:${contentHash}:${optionsHash}`;
  }

  /**
   * Check if entry is expired
   */
  private isExpired(entry: CachedDiagram): boolean {
    return Date.now() - entry.timestamp > this.config.ttl;
  }

  /**
   * Evict least recently used entries
   */
  private evictLRU(): void {
    if (this.cache.size <= this.config.maxEntries && this.currentSize <= this.config.maxSize) {
      return;
    }

    // Sort by access count and last accessed time
    const entries = Array.from(this.cache.entries()).sort(([, a], [, b]) => {
      if (a.accessCount !== b.accessCount) {
        return a.accessCount - b.accessCount;
      }
      return a.lastAccessed - b.lastAccessed;
    });

    // Remove entries until we're under limits
    while (
      (this.cache.size > this.config.maxEntries || this.currentSize > this.config.maxSize) &&
      entries.length > 0
    ) {
      const [key, entry] = entries.shift()!;
      this.cache.delete(key);
      this.currentSize -= entry.size;
    }
  }

  /**
   * Clean up expired entries
   */
  private cleanup(): void {
    for (const [key, entry] of this.cache.entries()) {
      if (this.isExpired(entry)) {
        this.cache.delete(key);
        this.currentSize -= entry.size;
      }
    }
  }

  /**
   * Get cached diagram
   */
  get(content: string, type: string, theme?: string, options?: any): any | null {
    this.cleanup();
    const key = this.generateKey(content, type, theme, options);
    const entry = this.cache.get(key);
    
    if (!entry || this.isExpired(entry)) {
      if (entry) {
        this.cache.delete(key);
        this.currentSize -= entry.size;
      }
      return null;
    }

    // Update access statistics
    entry.accessCount++;
    entry.lastAccessed = Date.now();
    
    return entry.content;
  }

  /**
   * Set cached diagram
   */
  set(content: string, type: string, renderedContent: any, theme?: string, options?: any): void {
    this.cleanup();
    this.evictLRU();
    
    const key = this.generateKey(content, type, theme, options);
    const size = estimateSize(renderedContent);
    
    const entry: CachedDiagram = {
      content: renderedContent,
      timestamp: Date.now(),
      accessCount: 1,
      lastAccessed: Date.now(),
      size,
    };

    // Remove existing entry if present
    const existing = this.cache.get(key);
    if (existing) {
      this.currentSize -= existing.size;
    }

    this.cache.set(key, entry);
    this.currentSize += size;
  }

  /**
   * Check if diagram is cached
   */
  has(content: string, type: string, theme?: string, options?: any): boolean {
    this.cleanup();
    const key = this.generateKey(content, type, theme, options);
    const entry = this.cache.get(key);
    return entry !== undefined && !this.isExpired(entry);
  }

  /**
   * Clear all cached diagrams
   */
  clear(): void {
    this.cache.clear();
    this.currentSize = 0;
  }

  /**
   * Get cache statistics
   */
  getStats(): { entries: number; size: number; maxEntries: number; maxSize: number } {
    this.cleanup();
    return {
      entries: this.cache.size,
      size: this.currentSize,
      maxEntries: this.config.maxEntries,
      maxSize: this.config.maxSize,
    };
  }
}

/**
 * Default diagram cache instance
 */
export const diagramCache = new DiagramCache();

/**
 * Export the class for custom instances
 */
export { DiagramCache };
export type { CacheConfig, CachedDiagram };