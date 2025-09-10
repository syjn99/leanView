import { useQuery } from '@connectrpc/connect-query';
import type { FC } from 'react';

import { getAllClientsHeads } from '../gen/proto/api/v1/monitoring-MonitoringService_connectquery';

export const Example: FC = () => {
    const { data, isLoading, error } = useQuery(getAllClientsHeads, {});

    if (isLoading) {
        return (
            <div className='p-4'>
                <div className='text-gray-600'>Loading client data...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className='p-4'>
                <div className='text-red-600'>Error: {error.message}</div>
            </div>
        );
    }

    if (!data) {
        return (
            <div className='p-4'>
                <div className='text-gray-600'>No data available</div>
            </div>
        );
    }

    return (
        <div className='p-4'>
            <h2 className='mb-4 text-xl font-bold'>
                Client Monitoring Dashboard
            </h2>

            <div className='mb-4 rounded bg-gray-100 p-3'>
                <div className='text-sm text-gray-600'>
                    Total Clients: {data.totalClients} | Healthy:{' '}
                    {data.healthyClients}
                </div>
            </div>

            <div className='space-y-4'>
                {data.clientHeads.map((client, index) => (
                    <div
                        key={client.clientLabel || index}
                        className={`rounded-lg border p-4 ${
                            client.isHealthy
                                ? 'border-green-200 bg-green-50'
                                : 'border-red-200 bg-red-50'
                        }`}
                    >
                        <div className='mb-2 flex items-center justify-between'>
                            <h3 className='font-semibold'>
                                {client.clientLabel}
                            </h3>
                            <span
                                className={`rounded px-2 py-1 text-xs ${
                                    client.isHealthy
                                        ? 'bg-green-200 text-green-800'
                                        : 'bg-red-200 text-red-800'
                                }`}
                            >
                                {client.isHealthy ? 'Healthy' : 'Unhealthy'}
                            </span>
                        </div>

                        <div className='mb-2 text-sm text-gray-600'>
                            <div>Endpoint: {client.endpointUrl}</div>
                            <div>
                                Last Update:{' '}
                                {new Date(
                                    Number(client.lastUpdateMs),
                                ).toLocaleString()}
                            </div>
                        </div>

                        {client.blockHeader && (
                            <div className='rounded border bg-white p-3'>
                                <h4 className='mb-2 font-medium'>
                                    Block Header
                                </h4>
                                <div className='space-y-1 text-xs'>
                                    <div>
                                        Slot:{' '}
                                        {client.blockHeader.slot.toString()}
                                    </div>
                                    <div>
                                        Proposer Index:{' '}
                                        {client.blockHeader.proposerIndex.toString()}
                                    </div>
                                    <div>Block Root: {client.blockRoot}</div>
                                    <div>
                                        Parent Root:{' '}
                                        {client.blockHeader.parentRoot}
                                    </div>
                                    <div>
                                        State Root:{' '}
                                        {client.blockHeader.stateRoot}
                                    </div>
                                    <div>
                                        Body Root: {client.blockHeader.bodyRoot}
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
};
