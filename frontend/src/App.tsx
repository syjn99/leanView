import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { ClientMonitoringDashboard } from './components/ClientMonitoringDashboard';

const finalTransport = createConnectTransport({
    baseUrl: 'http://localhost:8080',
});

const queryClient = new QueryClient();

function App() {
    return (
        <TransportProvider transport={finalTransport}>
            <QueryClientProvider client={queryClient}>
                <div className='min-h-screen bg-background'>
                    <main className='min-h-screen'>
                        <ClientMonitoringDashboard />
                    </main>
                </div>
            </QueryClientProvider>
        </TransportProvider>
    );
}

export default App;
