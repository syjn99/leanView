import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { Example } from './components/Example';

const finalTransport = createConnectTransport({
    baseUrl: 'http://localhost:8080',
});

const queryClient = new QueryClient();

function App() {
    return (
        <TransportProvider transport={finalTransport}>
            <QueryClientProvider client={queryClient}>
                <div className='min-h-screen bg-gray-100'>
                    <header className='border-b border-gray-200 bg-white shadow-sm'>
                        <div className='mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8'>
                            <h1 className='text-3xl font-bold text-gray-900'>
                                Lean View - PQ Devnet Visualizer
                            </h1>
                            <p className='mt-2 text-gray-600'>
                                Monitoring PQ Devnet blockchain clients in
                                real-time
                            </p>
                        </div>
                    </header>

                    <main>
                        <Example />
                    </main>
                </div>
            </QueryClientProvider>
        </TransportProvider>
    );
}

export default App;
