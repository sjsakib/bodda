// Type declarations for diagram libraries

declare module 'mermaid' {
  interface MermaidConfig {
    theme?: 'default' | 'dark' | 'forest' | 'neutral';
    themeVariables?: {
      primaryColor?: string;
      primaryTextColor?: string;
      primaryBorderColor?: string;
      lineColor?: string;
      backgroundColor?: string;
    };
    fontFamily?: string;
    fontSize?: number;
    securityLevel?: 'strict' | 'loose';
    maxTextSize?: number;
  }

  interface RenderResult {
    svg: string;
    bindFunctions?: (element: Element) => void;
  }

  const mermaid: {
    initialize: (config: MermaidConfig) => void;
    render: (id: string, definition: string) => Promise<RenderResult>;
    parse: (definition: string) => Promise<boolean>;
  };

  export default mermaid;
}

declare module 'react-vega' {
  import { ComponentType } from 'react';
  
  interface VegaLiteProps {
    spec: any;
    data?: any;
    width?: number;
    height?: number;
    actions?: boolean;
    renderer?: 'canvas' | 'svg';
    className?: string;
    onError?: (error: Error) => void;
    onParseError?: (error: Error) => void;
  }

  export const VegaLite: ComponentType<VegaLiteProps>;
}

declare module 'vega-lite' {
  export interface TopLevelSpec {
    $schema?: string;
    data?: any;
    mark?: string | object;
    encoding?: any;
    transform?: any[];
    config?: any;
    width?: number;
    height?: number;
    title?: string | object;
    description?: string;
  }

  export function compile(spec: TopLevelSpec): any;
}