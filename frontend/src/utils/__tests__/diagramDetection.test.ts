import {
  detectDiagrams,
  hasDiagramContent,
  extractMermaidDiagrams,
  extractVegaLiteDiagrams,
  validateMermaidSyntax,
  validateVegaLiteSpec,
  sanitizeDiagramContent,
  countDiagrams,
  MERMAID_PATTERN,
  VEGA_LITE_PATTERN,
} from '../diagramDetection';

describe('Diagram Detection Utilities', () => {
  describe('Pattern Matching', () => {
    test('MERMAID_PATTERN matches valid mermaid code blocks', () => {
      const content = '```mermaid\ngraph TD\nA-->B\n```';
      const matches = content.match(MERMAID_PATTERN);
      
      expect(matches).toHaveLength(1);
      expect(matches![0]).toBe('```mermaid\ngraph TD\nA-->B\n```');
    });

    test('VEGA_LITE_PATTERN matches valid vega-lite code blocks', () => {
      const content = '```vega-lite\n{"mark": "bar"}\n```';
      const matches = content.match(VEGA_LITE_PATTERN);
      
      expect(matches).toHaveLength(1);
      expect(matches![0]).toBe('```vega-lite\n{"mark": "bar"}\n```');
    });

    test('patterns handle whitespace variations', () => {
      const mermaidContent = '```mermaid   \n\ngraph TD\nA-->B\n\n```';
      const vegaContent = '```vega-lite\n\n{"mark": "bar"}\n\n```';
      
      expect(MERMAID_PATTERN.test(mermaidContent)).toBe(true);
      expect(VEGA_LITE_PATTERN.test(vegaContent)).toBe(true);
    });
  });

  describe('detectDiagrams', () => {
    test('detects single mermaid diagram', () => {
      const content = 'Some text\n\n```mermaid\ngraph TD\nA-->B\n```\n\nMore text';
      const result = detectDiagrams(content);

      expect(result.hasDiagrams).toBe(true);
      expect(result.mermaidCount).toBe(1);
      expect(result.vegaLiteCount).toBe(0);
      expect(result.totalCount).toBe(1);
      expect(result.diagrams).toHaveLength(1);
      expect(result.diagrams[0].type).toBe('mermaid');
      expect(result.diagrams[0].content).toBe('graph TD\nA-->B');
    });

    test('detects single vega-lite diagram', () => {
      const content = 'Chart:\n\n```vega-lite\n{"mark": "bar", "data": {"values": []}}\n```';
      const result = detectDiagrams(content);

      expect(result.hasDiagrams).toBe(true);
      expect(result.mermaidCount).toBe(0);
      expect(result.vegaLiteCount).toBe(1);
      expect(result.totalCount).toBe(1);
      expect(result.diagrams[0].type).toBe('vega-lite');
      expect(result.diagrams[0].content).toBe('{"mark": "bar", "data": {"values": []}}');
    });

    test('detects multiple diagrams of different types', () => {
      const content = [
        '# Training Plan',
        '',
        '```mermaid',
        'graph TD',
        'A[Start] --> B[Warm up]',
        'B --> C[Main workout]',
        '```',
        '',
        '## Progress Chart',
        '',
        '```vega-lite',
        '{',
        '  "mark": "line",',
        '  "data": {"values": [{"x": 1, "y": 2}]}',
        '}',
        '```',
        '',
        'Another diagram:',
        '',
        '```mermaid',
        'sequenceDiagram',
        'Alice->>Bob: Hello',
        '```'
      ].join('\n');

      const result = detectDiagrams(content);

      expect(result.hasDiagrams).toBe(true);
      expect(result.mermaidCount).toBe(2);
      expect(result.vegaLiteCount).toBe(1);
      expect(result.totalCount).toBe(3);
      expect(result.diagrams).toHaveLength(3);
      
      // Check order is preserved
      expect(result.diagrams[0].type).toBe('mermaid');
      expect(result.diagrams[1].type).toBe('vega-lite');
      expect(result.diagrams[2].type).toBe('mermaid');
    });

    test('returns empty result for content without diagrams', () => {
      const content = 'Just regular markdown content with no diagrams';
      const result = detectDiagrams(content);

      expect(result.hasDiagrams).toBe(false);
      expect(result.mermaidCount).toBe(0);
      expect(result.vegaLiteCount).toBe(0);
      expect(result.totalCount).toBe(0);
      expect(result.diagrams).toHaveLength(0);
    });

    test('handles empty content', () => {
      const result = detectDiagrams('');

      expect(result.hasDiagrams).toBe(false);
      expect(result.totalCount).toBe(0);
    });
  });

  describe('hasDiagramContent', () => {
    test('returns true for content with mermaid diagrams', () => {
      const content = 'Text\n```mermaid\ngraph TD\nA-->B\n```';
      expect(hasDiagramContent(content)).toBe(true);
    });

    test('returns true for content with vega-lite diagrams', () => {
      const content = 'Text\n```vega-lite\n{"mark": "bar"}\n```';
      expect(hasDiagramContent(content)).toBe(true);
    });

    test('returns false for content without diagrams', () => {
      const content = 'Just regular text and ```code``` blocks';
      expect(hasDiagramContent(content)).toBe(false);
    });
  });

  describe('extractMermaidDiagrams', () => {
    test('extracts multiple mermaid diagrams', () => {
      const content = [
        '```mermaid',
        'graph TD',
        'A-->B',
        '```',
        '',
        'Some text',
        '',
        '```mermaid',
        'sequenceDiagram',
        'Alice->>Bob: Hello',
        '```'
      ].join('\n');

      const diagrams = extractMermaidDiagrams(content);
      
      expect(diagrams).toHaveLength(2);
      expect(diagrams[0]).toBe('graph TD\nA-->B');
      expect(diagrams[1]).toBe('sequenceDiagram\nAlice->>Bob: Hello');
    });

    test('returns empty array for content without mermaid diagrams', () => {
      const content = 'No diagrams here\n```vega-lite\n{"mark": "bar"}\n```';
      const diagrams = extractMermaidDiagrams(content);
      
      expect(diagrams).toHaveLength(0);
    });
  });

  describe('extractVegaLiteDiagrams', () => {
    test('extracts multiple vega-lite diagrams', () => {
      const content = [
        '```vega-lite',
        '{"mark": "bar"}',
        '```',
        '',
        'Text',
        '',
        '```vega-lite',
        '{"mark": "line", "data": {"values": []}}',
        '```'
      ].join('\n');

      const diagrams = extractVegaLiteDiagrams(content);
      
      expect(diagrams).toHaveLength(2);
      expect(diagrams[0]).toBe('{"mark": "bar"}');
      expect(diagrams[1]).toBe('{"mark": "line", "data": {"values": []}}');
    });

    test('returns empty array for content without vega-lite diagrams', () => {
      const content = 'No charts\n```mermaid\ngraph TD\nA-->B\n```';
      const diagrams = extractVegaLiteDiagrams(content);
      
      expect(diagrams).toHaveLength(0);
    });
  });

  describe('validateMermaidSyntax', () => {
    test('validates flowchart syntax', () => {
      const content = 'graph TD\nA[Start] --> B[End]';
      const result = validateMermaidSyntax(content);

      expect(result.isValid).toBe(true);
      expect(result.diagramType).toBe('flowchart');
      expect(result.errors).toHaveLength(0);
    });

    test('validates sequence diagram syntax', () => {
      const content = 'sequenceDiagram\nAlice->>Bob: Hello';
      const result = validateMermaidSyntax(content);

      expect(result.isValid).toBe(true);
      expect(result.diagramType).toBe('sequence');
      expect(result.errors).toHaveLength(0);
    });

    test('validates class diagram syntax', () => {
      const content = 'classDiagram\nclass Animal';
      const result = validateMermaidSyntax(content);

      expect(result.isValid).toBe(true);
      expect(result.diagramType).toBe('classDiagram');
      expect(result.errors).toHaveLength(0);
    });

    test('detects empty content', () => {
      const result = validateMermaidSyntax('');

      expect(result.isValid).toBe(false);
      expect(result.diagramType).toBe(null);
      expect(result.errors).toContain('Empty diagram content');
    });

    test('detects unrecognized diagram type', () => {
      const content = 'unknown diagram type\nsome content';
      const result = validateMermaidSyntax(content);

      expect(result.isValid).toBe(false);
      expect(result.diagramType).toBe(null);
      expect(result.errors).toContain('Unrecognized Mermaid diagram type');
    });

    test('detects unmatched brackets', () => {
      const content = 'graph TD\nA[Start --> B[End]';
      const result = validateMermaidSyntax(content);

      expect(result.isValid).toBe(false);
      expect(result.errors.some(error => error.includes('Unmatched brackets'))).toBe(true);
    });
  });

  describe('validateVegaLiteSpec', () => {
    test('validates basic bar chart spec', () => {
      const content = '{"mark": "bar", "data": {"values": []}}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(true);
      expect(result.spec).toEqual({"mark": "bar", "data": {"values": []}});
      expect(result.errors).toHaveLength(0);
    });

    test('validates line chart spec', () => {
      const content = '{"mark": "line", "encoding": {"x": {"field": "x"}}}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(true);
      expect(result.spec.mark).toBe('line');
    });

    test('detects invalid JSON', () => {
      const content = '{invalid json}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(false);
      expect(result.spec).toBe(null);
      expect(result.errors.some(error => error.includes('Invalid JSON'))).toBe(true);
    });

    test('detects empty content', () => {
      const result = validateVegaLiteSpec('');

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Empty specification content');
    });

    test('detects missing required properties', () => {
      const content = '{"data": {"values": []}}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(false);
      expect(result.errors.some(error => 
        error.includes('must have a mark, layer, concat, facet, or repeat property')
      )).toBe(true);
    });

    test('detects invalid mark type', () => {
      const content = '{"mark": "invalid-mark", "data": {"values": []}}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(false);
      expect(result.errors.some(error => error.includes('Invalid mark type'))).toBe(true);
    });

    test('detects dangerous properties', () => {
      const content = '{"mark": "bar", "datasets": {"external": "data"}}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(false);
      expect(result.errors.some(error => error.includes("Property 'datasets' is not allowed"))).toBe(true);
    });

    test('validates complex mark object', () => {
      const content = '{"mark": {"type": "bar", "color": "blue"}, "data": {"values": []}}';
      const result = validateVegaLiteSpec(content);

      expect(result.isValid).toBe(true);
      expect(result.spec.mark.type).toBe('bar');
    });
  });

  describe('sanitizeDiagramContent', () => {
    test('sanitizes mermaid content', () => {
      const content = [
        '%%{config: {"theme": "dark"}}%%',
        'graph TD',
        'A[Start] --> B[End]',
        'click A href "javascript:alert(\'xss\')"'
      ].join('\n');

      const sanitized = sanitizeDiagramContent(content, 'mermaid');

      expect(sanitized).not.toContain('%%{config');
      expect(sanitized).not.toContain('click A href');
      expect(sanitized).not.toContain('javascript:');
      expect(sanitized).toContain('graph TD');
      expect(sanitized).toContain('A[Start] --> B[End]');
    });

    test('sanitizes vega-lite content', () => {
      const content = JSON.stringify({
        mark: 'bar',
        data: { values: [] },
        datasets: { external: 'dangerous' },
        transform: [{ filter: 'dangerous' }]
      });

      const sanitized = sanitizeDiagramContent(content, 'vega-lite');
      const parsed = JSON.parse(sanitized);

      expect(parsed.mark).toBe('bar');
      expect(parsed.data).toEqual({ values: [] });
      expect(parsed.datasets).toBeUndefined();
      expect(parsed.transform).toBeUndefined();
    });

    test('handles invalid vega-lite JSON gracefully', () => {
      const content = '{invalid json}';
      const sanitized = sanitizeDiagramContent(content, 'vega-lite');

      expect(sanitized).toBe(content); // Returns original if parsing fails
    });
  });

  describe('countDiagrams', () => {
    test('counts diagrams correctly', () => {
      const content = [
        '```mermaid',
        'graph TD',
        'A-->B',
        '```',
        '',
        '```vega-lite',
        '{"mark": "bar"}',
        '```',
        '',
        '```mermaid',
        'sequenceDiagram',
        'Alice->>Bob: Hello',
        '```'
      ].join('\n');

      const counts = countDiagrams(content);

      expect(counts.total).toBe(3);
      expect(counts.mermaid).toBe(2);
      expect(counts.vegaLite).toBe(1);
    });

    test('returns zero counts for content without diagrams', () => {
      const content = 'Just regular markdown content';
      const counts = countDiagrams(content);

      expect(counts.total).toBe(0);
      expect(counts.mermaid).toBe(0);
      expect(counts.vegaLite).toBe(0);
    });
  });

  describe('Edge Cases', () => {
    test('handles nested code blocks', () => {
      const content = [
        '```markdown',
        'Here\'s how to write a mermaid diagram:',
        '',
        '```mermaid',
        'graph TD',
        'A-->B',
        '```',
        '```',
        '',
        '```mermaid',
        'graph LR',
        'X-->Y',
        '```'
      ].join('\n');

      const result = detectDiagrams(content);
      
      // Our simple regex approach will detect both mermaid blocks
      // This is acceptable behavior for the current implementation
      expect(result.mermaidCount).toBe(2);
      expect(result.diagrams[0].content).toBe('graph TD\nA-->B');
      expect(result.diagrams[1].content).toBe('graph LR\nX-->Y');
    });

    test('handles diagrams with special characters', () => {
      const mermaidContent = [
        '```mermaid',
        'graph TD',
        'A["Special chars: !@#$%^&*()"] --> B',
        '```'
      ].join('\n');

      const vegaContent = [
        '```vega-lite',
        '{',
        '  "mark": "text",',
        '  "data": {"values": [{"text": "Special: !@#$%^&*()"}]}',
        '}',
        '```'
      ].join('\n');

      expect(hasDiagramContent(mermaidContent)).toBe(true);
      expect(hasDiagramContent(vegaContent)).toBe(true);
    });

    test('handles very large diagram content', () => {
      const largeContent = 'A'.repeat(10000);
      const content = ['```mermaid', 'graph TD', largeContent, '```'].join('\n');

      const result = detectDiagrams(content);
      
      expect(result.hasDiagrams).toBe(true);
      expect(result.diagrams[0].content).toContain(largeContent);
    });
  });
});