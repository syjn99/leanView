import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { ClientMonitoringDashboard } from './components/ClientMonitoringDashboard';
import { BlockHeadersTable } from './components/BlockHeadersTable';

const finalTransport = createConnectTransport({
    baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
});

const queryClient = new QueryClient();

function App() {
    return (
        <TransportProvider transport={finalTransport}>
            <QueryClientProvider client={queryClient}>
                <div className='min-h-screen bg-background'>
                    <main className='min-h-screen'>
                        <div className='container mx-auto p-6 space-y-6'>
                            <ClientMonitoringDashboard />
                            <BlockHeadersTable />
                        </div>
                    </main>
                </div>
            </QueryClientProvider>
        </TransportProvider>
    );
}

export default App;
