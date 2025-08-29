import { useEffect, useMemo } from 'react';
import { useDiagramLibrary } from '../contexts/DiagramLibraryContext';
import { detectDiagrams, hasDiagramContent } from '../utils/diagramDetection';

/**
 * Interface for the diagram loader hook return value
 */
interface UseDiagramLoaderReturn {
  /** Whether the content has any diagrams */
  hasDiagrams: boolean;
  /** Detailed diagram detection results */
  diagramInfo: ReturnType<typeof detectDiagrams>;
  /** Current library loading states */
  libraryState: ReturnType<typeof useDiagramLibrary>['libraryState'];
  /** Whether any libraries are currently loading */
  isLoading: boolean;
  /** Whether all required libraries are loaded */
  allRequiredLibrariesLoaded: boolean;
  /** Any errors from library loading */
  errors: string[];
  /** Function to manually trigger library loading */
  loadRequiredLibraries: () => Promise<void>;
}

/**
 * Custom hook that detects diagrams in content and manages library loading
 * Automatically loads required diagram libraries when diagram content is detected
 * 
 * @param content - The markdown content to analyze for diagrams
 * @param autoLoad - Whether to automatically load libraries when diagrams are detected (default: true)
 * @returns Object with diagram detection results and library loading state
 */
export const useDiagramLoader = (
  content: string, 
  autoLoad: boolean = true
): UseDiagramLoaderReturn => {
  const { libraryState, loadMermaid, loadVegaLite } = useDiagramLibrary();

  // Detect diagrams in the content
  const diagramInfo = useMemo(() => {
    return detectDiagrams(content);
  }, [content]);

  const hasDiagrams = useMemo(() => {
    return hasDiagramContent(content);
  }, [content]);

  // Determine which libraries are needed
  const needsMermaid = diagramInfo.mermaidCount > 0;
  const needsVegaLite = diagramInfo.vegaLiteCount > 0;

  // Check loading states
  const isLoading = useMemo(() => {
    return (needsMermaid && libraryState.mermaid.loading) || 
           (needsVegaLite && libraryState.vegaLite.loading);
  }, [needsMermaid, needsVegaLite, libraryState]);

  // Check if all required libraries are loaded
  const allRequiredLibrariesLoaded = useMemo(() => {
    const mermaidReady = !needsMermaid || libraryState.mermaid.loaded;
    const vegaLiteReady = !needsVegaLite || libraryState.vegaLite.loaded;
    return mermaidReady && vegaLiteReady;
  }, [needsMermaid, needsVegaLite, libraryState]);

  // Collect any errors
  const errors = useMemo(() => {
    const errorList: string[] = [];
    
    if (needsMermaid && libraryState.mermaid.error) {
      errorList.push(`Mermaid: ${libraryState.mermaid.error}`);
    }
    
    if (needsVegaLite && libraryState.vegaLite.error) {
      errorList.push(`Vega-Lite: ${libraryState.vegaLite.error}`);
    }
    
    return errorList;
  }, [needsMermaid, needsVegaLite, libraryState]);

  // Function to manually load required libraries
  const loadRequiredLibraries = async () => {
    const promises: Promise<void>[] = [];
    
    if (needsMermaid && !libraryState.mermaid.loaded && !libraryState.mermaid.loading) {
      promises.push(loadMermaid());
    }
    
    if (needsVegaLite && !libraryState.vegaLite.loaded && !libraryState.vegaLite.loading) {
      promises.push(loadVegaLite());
    }
    
    if (promises.length > 0) {
      await Promise.allSettled(promises);
    }
  };

  // Auto-load libraries when diagrams are detected
  useEffect(() => {
    if (autoLoad && hasDiagrams && !allRequiredLibrariesLoaded && !isLoading) {
      loadRequiredLibraries().catch(error => {
        console.error('Failed to auto-load diagram libraries:', error);
      });
    }
  }, [autoLoad, hasDiagrams, allRequiredLibrariesLoaded, isLoading]);

  return {
    hasDiagrams,
    diagramInfo,
    libraryState,
    isLoading,
    allRequiredLibrariesLoaded,
    errors,
    loadRequiredLibraries,
  };
};

/**
 * Simplified hook that just checks if content has diagrams and loads libraries
 * Useful for components that only need to know if libraries should be loaded
 * 
 * @param content - The markdown content to analyze
 * @returns Boolean indicating if diagrams are present and libraries are being loaded
 */
export const useHasDiagrams = (content: string): boolean => {
  const { hasDiagrams } = useDiagramLoader(content, true);
  return hasDiagrams;
};

/**
 * Hook that provides loading state for specific diagram types
 * Useful for showing type-specific loading indicators
 * 
 * @param content - The markdown content to analyze
 * @returns Object with loading states for each diagram type
 */
export const useDiagramLoadingStates = (content: string) => {
  const { diagramInfo, libraryState, isLoading } = useDiagramLoader(content, true);
  
  return {
    mermaid: {
      hasContent: diagramInfo.mermaidCount > 0,
      isLoading: diagramInfo.mermaidCount > 0 && libraryState.mermaid.loading,
      isLoaded: libraryState.mermaid.loaded,
      error: libraryState.mermaid.error,
    },
    vegaLite: {
      hasContent: diagramInfo.vegaLiteCount > 0,
      isLoading: diagramInfo.vegaLiteCount > 0 && libraryState.vegaLite.loading,
      isLoaded: libraryState.vegaLite.loaded,
      error: libraryState.vegaLite.error,
    },
    overall: {
      isLoading,
      hasAnyDiagrams: diagramInfo.hasDiagrams,
      totalCount: diagramInfo.totalCount,
    }
  };
};