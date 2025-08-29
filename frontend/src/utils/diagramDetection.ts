/**
 * Diagram Detection Utilities
 * 
 * This module provides utilities for detecting and validating diagram content
 * in markdown text, specifically for Mermaid and Vega-Lite diagrams.
 */

// Regex patterns for detecting diagram code blocks
export const MERMAID_PATTERN = /```mermaid\s*\n([\s\S]*?)\n```/g;
export const VEGA_LITE_PATTERN = /```vega-lite\s*\n([\s\S]*?)\n```/g;

// More specific patterns for validation
const MERMAID_SYNTAX_PATTERNS = {
  flowchart: /^\s*(graph|flowchart)\s+(TD|TB|BT|RL|LR)/m,
  sequence: /^\s*sequenceDiagram/m,
  classDiagram: /^\s*classDiagram/m,
  stateDiagram: /^\s*stateDiagram(-v2)?/m,
  erDiagram: /^\s*erDiagram/m,
  journey: /^\s*journey/m,
  gantt: /^\s*gantt/m,
  pie: /^\s*pie(\s+title\s+.+)?/m,
  gitgraph: /^\s*gitgraph/m,
  mindmap: /^\s*mindmap/m,
  timeline: /^\s*timeline/m,
};

export interface DiagramMatch {
  type: 'mermaid' | 'vega-lite';
  content: string;
  startIndex: number;
  endIndex: number;
  fullMatch: string;
}

export interface DiagramDetectionResult {
  hasDiagrams: boolean;
  mermaidCount: number;
  vegaLiteCount: number;
  totalCount: number;
  diagrams: DiagramMatch[];
}

/**
 * Detects all diagram code blocks in markdown content
 */
export function detectDiagrams(content: string): DiagramDetectionResult {
  const diagrams: DiagramMatch[] = [];
  let mermaidCount = 0;
  let vegaLiteCount = 0;

  // Reset regex lastIndex to ensure fresh matching
  MERMAID_PATTERN.lastIndex = 0;
  VEGA_LITE_PATTERN.lastIndex = 0;

  // Find all Mermaid diagrams
  let mermaidMatch;
  while ((mermaidMatch = MERMAID_PATTERN.exec(content)) !== null) {
    diagrams.push({
      type: 'mermaid',
      content: mermaidMatch[1].trim(),
      startIndex: mermaidMatch.index,
      endIndex: mermaidMatch.index + mermaidMatch[0].length,
      fullMatch: mermaidMatch[0],
    });
    mermaidCount++;
  }

  // Find all Vega-Lite diagrams
  let vegaMatch;
  while ((vegaMatch = VEGA_LITE_PATTERN.exec(content)) !== null) {
    diagrams.push({
      type: 'vega-lite',
      content: vegaMatch[1].trim(),
      startIndex: vegaMatch.index,
      endIndex: vegaMatch.index + vegaMatch[0].length,
      fullMatch: vegaMatch[0],
    });
    vegaLiteCount++;
  }

  // Sort diagrams by their position in the content
  diagrams.sort((a, b) => a.startIndex - b.startIndex);

  return {
    hasDiagrams: diagrams.length > 0,
    mermaidCount,
    vegaLiteCount,
    totalCount: diagrams.length,
    diagrams,
  };
}

/**
 * Checks if content contains any diagram code blocks
 */
export function hasDiagramContent(content: string): boolean {
  return MERMAID_PATTERN.test(content) || VEGA_LITE_PATTERN.test(content);
}

/**
 * Extracts only Mermaid diagram content from markdown
 */
export function extractMermaidDiagrams(content: string): string[] {
  const diagrams: string[] = [];
  MERMAID_PATTERN.lastIndex = 0;
  
  let match;
  while ((match = MERMAID_PATTERN.exec(content)) !== null) {
    diagrams.push(match[1].trim());
  }
  
  return diagrams;
}

/**
 * Extracts only Vega-Lite diagram content from markdown
 */
export function extractVegaLiteDiagrams(content: string): string[] {
  const diagrams: string[] = [];
  VEGA_LITE_PATTERN.lastIndex = 0;
  
  let match;
  while ((match = VEGA_LITE_PATTERN.exec(content)) !== null) {
    diagrams.push(match[1].trim());
  }
  
  return diagrams;
}

/**
 * Validates Mermaid diagram syntax
 */
export function validateMermaidSyntax(content: string): {
  isValid: boolean;
  diagramType: string | null;
  errors: string[];
} {
  const errors: string[] = [];
  let diagramType: string | null = null;
  
  if (!content || content.trim().length === 0) {
    errors.push('Empty diagram content');
    return { isValid: false, diagramType, errors };
  }

  // Check for recognized diagram types
  for (const [type, pattern] of Object.entries(MERMAID_SYNTAX_PATTERNS)) {
    if (pattern.test(content)) {
      diagramType = type;
      break;
    }
  }

  if (!diagramType) {
    errors.push('Unrecognized Mermaid diagram type');
  }

  // Basic syntax validation
  const lines = content.split('\n').map(line => line.trim()).filter(line => line.length > 0);
  
  if (lines.length === 0) {
    errors.push('No content lines found');
  }

  // Check for common syntax issues
  const hasUnmatchedBrackets = (str: string) => {
    const brackets = { '[': ']', '(': ')', '{': '}' };
    const stack: string[] = [];
    
    for (const char of str) {
      if (char in brackets) {
        stack.push(brackets[char as keyof typeof brackets]);
      } else if (Object.values(brackets).includes(char)) {
        if (stack.pop() !== char) return true;
      }
    }
    
    return stack.length > 0;
  };

  for (const line of lines) {
    if (hasUnmatchedBrackets(line)) {
      errors.push(`Unmatched brackets in line: ${line.substring(0, 50)}...`);
    }
  }

  return {
    isValid: errors.length === 0,
    diagramType,
    errors,
  };
}

/**
 * Validates Vega-Lite JSON specification
 */
export function validateVegaLiteSpec(content: string): {
  isValid: boolean;
  spec: any | null;
  errors: string[];
} {
  const errors: string[] = [];
  let spec: any = null;

  if (!content || content.trim().length === 0) {
    errors.push('Empty specification content');
    return { isValid: false, spec, errors };
  }

  try {
    spec = JSON.parse(content);
  } catch (parseError) {
    errors.push(`Invalid JSON: ${parseError instanceof Error ? parseError.message : 'Unknown parsing error'}`);
    return { isValid: false, spec, errors };
  }

  // Basic Vega-Lite validation
  if (typeof spec !== 'object' || spec === null) {
    errors.push('Specification must be a JSON object');
    return { isValid: false, spec, errors };
  }

  // Check for required properties
  if (!spec.mark && !spec.layer && !spec.concat && !spec.facet && !spec.repeat) {
    errors.push('Specification must have a mark, layer, concat, facet, or repeat property');
  }

  // Validate mark types if present
  if (spec.mark) {
    const validMarks = [
      'arc', 'area', 'bar', 'circle', 'line', 'point', 'rect', 'rule', 'square', 'text', 'tick', 'trail'
    ];
    const markType = typeof spec.mark === 'string' ? spec.mark : spec.mark?.type;
    
    if (markType && !validMarks.includes(markType)) {
      errors.push(`Invalid mark type: ${markType}`);
    }
  }

  // Check for potentially dangerous properties
  const dangerousProps = ['datasets', 'transform'];
  for (const prop of dangerousProps) {
    if (spec[prop]) {
      errors.push(`Property '${prop}' is not allowed for security reasons`);
    }
  }

  return {
    isValid: errors.length === 0,
    spec,
    errors,
  };
}

/**
 * Sanitizes diagram content for security
 */
export function sanitizeDiagramContent(content: string, type: 'mermaid' | 'vega-lite'): string {
  if (type === 'mermaid') {
    return content
      .replace(/%%\{.*?\}%%/g, '') // Remove config directives
      .replace(/click\s+\w+\s+href/gi, '') // Remove click handlers
      .replace(/javascript:/gi, '') // Remove javascript: URLs
      .trim();
  }
  
  if (type === 'vega-lite') {
    try {
      const spec = JSON.parse(content);
      // Remove potentially dangerous properties
      const sanitized = { ...spec };
      delete sanitized.datasets;
      delete sanitized.transform;
      return JSON.stringify(sanitized, null, 2);
    } catch {
      return content; // Return original if parsing fails
    }
  }
  
  return content;
}

/**
 * Counts the total number of diagrams in content
 */
export function countDiagrams(content: string): {
  total: number;
  mermaid: number;
  vegaLite: number;
} {
  const result = detectDiagrams(content);
  return {
    total: result.totalCount,
    mermaid: result.mermaidCount,
    vegaLite: result.vegaLiteCount,
  };
}