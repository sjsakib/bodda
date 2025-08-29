import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
} from 'react';

/**
 * State interface for individual diagram libraries
 */
interface LibraryState {
  loaded: boolean;
  loading: boolean;
  error: string | null;
  module?: any; // Store the actual imported module
}

/**
 * Complete state for all diagram libraries
 */
interface DiagramLibraryState {
  mermaid: LibraryState;
  vegaLite: LibraryState;
}

/**
 * Context value interface with state and loading functions
 */
interface DiagramLibraryContextValue {
  libraryState: DiagramLibraryState;
  loadMermaid: () => Promise<void>;
  loadVegaLite: () => Promise<void>;
  loadAllLibraries: () => Promise<void>;
}

/**
 * Initial state for all diagram libraries
 */
const initialState: DiagramLibraryState = {
  mermaid: { loaded: false, loading: false, error: null, module: null },
  vegaLite: { loaded: false, loading: false, error: null, module: null },
};

/**
 * React context for diagram library state management
 */
const DiagramLibraryContext = createContext<DiagramLibraryContextValue | undefined>(
  undefined
);

/**
 * Props interface for the DiagramLibraryProvider component
 */
interface DiagramLibraryProviderProps {
  children: ReactNode;
}

/**
 * Provider component that manages diagram library loading state and provides
 * dynamic import functions for mermaid and vega-lite libraries
 */
export const DiagramLibraryProvider: React.FC<DiagramLibraryProviderProps> = ({
  children,
}) => {
  const [libraryState, setLibraryState] = useState<DiagramLibraryState>(initialState);

  /**
   * Dynamically loads the Mermaid library with error handling
   */
  const loadMermaid = useCallback(async () => {
    // Skip if already loaded or currently loading
    if (libraryState.mermaid.loaded || libraryState.mermaid.loading) {
      return;
    }

    setLibraryState(prev => ({
      ...prev,
      mermaid: { ...prev.mermaid, loading: true, error: null },
    }));

    try {
      // Dynamic import with timeout
      const timeoutPromise = new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Mermaid library load timeout')), 10000)
      );

      const loadPromise = import('mermaid').then(async mermaidModule => {
        // Initialize mermaid with proper configuration
        const mermaid = mermaidModule.default;
        mermaid.initialize({
          theme: 'default',
          securityLevel: 'loose',
        });
        return mermaidModule;
      });

      const mermaidModule = await Promise.race([loadPromise, timeoutPromise]);

      setLibraryState(prev => ({
        ...prev,
        mermaid: { loaded: true, loading: false, error: null, module: mermaidModule },
      }));

      console.log('Mermaid library loaded successfully');
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to load Mermaid library';

      setLibraryState(prev => ({
        ...prev,
        mermaid: {
          loaded: false,
          loading: false,
          error: errorMessage,
          module: null,
        },
      }));

      console.error('Failed to load Mermaid library:', errorMessage);
    }
  }, [libraryState.mermaid]);

  /**
   * Dynamically loads the Vega-Lite and React-Vega libraries with error handling
   */
  const loadVegaLite = useCallback(async () => {
    // Skip if already loaded or currently loading
    if (libraryState.vegaLite.loaded || libraryState.vegaLite.loading) {
      return;
    }

    setLibraryState(prev => ({
      ...prev,
      vegaLite: { ...prev.vegaLite, loading: true, error: null },
    }));

    try {
      // Dynamic import with timeout for both libraries
      const timeoutPromise = new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Vega-Lite library load timeout')), 10000)
      );

      const loadPromise = Promise.all([import('vega-lite'), import('react-vega')]);

      const modules = (await Promise.race([loadPromise, timeoutPromise])) as [any, any];
      const [vegaLiteModule, reactVegaModule] = modules;

      setLibraryState(prev => ({
        ...prev,
        vegaLite: {
          loaded: true,
          loading: false,
          error: null,
          module: { vegaLite: vegaLiteModule, reactVega: reactVegaModule },
        },
      }));

      console.log('Vega-Lite libraries loaded successfully');
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to load Vega-Lite libraries';

      setLibraryState(prev => ({
        ...prev,
        vegaLite: {
          loaded: false,
          loading: false,
          error: errorMessage,
          module: null,
        },
      }));

      console.error('Failed to load Vega-Lite libraries:', errorMessage);
    }
  }, [libraryState.vegaLite]);

  /**
   * Loads all diagram libraries concurrently
   */
  const loadAllLibraries = useCallback(async () => {
    await Promise.allSettled([loadMermaid(), loadVegaLite()]);
  }, [loadMermaid, loadVegaLite]);

  const contextValue: DiagramLibraryContextValue = {
    libraryState,
    loadMermaid,
    loadVegaLite,
    loadAllLibraries,
  };

  return (
    <DiagramLibraryContext.Provider value={contextValue}>
      {children}
    </DiagramLibraryContext.Provider>
  );
};

/**
 * Custom hook to access diagram library context
 * @throws Error if used outside of DiagramLibraryProvider
 */
export const useDiagramLibrary = (): DiagramLibraryContextValue => {
  const context = useContext(DiagramLibraryContext);

  if (context === undefined) {
    throw new Error('useDiagramLibrary must be used within a DiagramLibraryProvider');
  }

  return context;
};

/**
 * Type exports for external use
 */
export type { DiagramLibraryState, LibraryState, DiagramLibraryContextValue };
