import { useQuery } from "@connectrpc/connect-query";
import type { FC } from "react";
import { getAllClientsHeads } from "../gen/proto/api/v1/monitoring-MonitoringService_connectquery";

export const Example: FC = () => {
    const { data, isLoading, error } = useQuery(getAllClientsHeads, {});

    if (isLoading) {
        return (
            <div className="p-4">
                <div className="text-gray-600">Loading client data...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-4">
                <div className="text-red-600">Error: {error.message}</div>
            </div>
        );
    }

    if (!data) {
        return (
            <div className="p-4">
                <div className="text-gray-600">No data available</div>
            </div>
        );
    }

    return (
        <div className="p-4">
            <h2 className="text-xl font-bold mb-4">Client Monitoring Dashboard</h2>

            <div className="mb-4 bg-gray-100 p-3 rounded">
                <div className="text-sm text-gray-600">
                    Total Clients: {data.totalClients} | Healthy: {data.healthyClients}
                </div>
            </div>

            <div className="space-y-4">
                {data.clientHeads.map((client, index) => (
                    <div
                        key={client.clientLabel || index}
                        className={`border rounded-lg p-4 ${client.isHealthy ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'
                            }`}
                    >
                        <div className="flex items-center justify-between mb-2">
                            <h3 className="font-semibold">{client.clientLabel}</h3>
                            <span
                                className={`px-2 py-1 rounded text-xs ${client.isHealthy
                                    ? 'bg-green-200 text-green-800'
                                    : 'bg-red-200 text-red-800'
                                    }`}
                            >
                                {client.isHealthy ? 'Healthy' : 'Unhealthy'}
                            </span>
                        </div>

                        <div className="text-sm text-gray-600 mb-2">
                            <div>Endpoint: {client.endpointUrl}</div>
                            <div>Last Update: {new Date(Number(client.lastUpdateMs)).toLocaleString()}</div>
                        </div>

                        {client.blockHeader && (
                            <div className="bg-white p-3 rounded border">
                                <h4 className="font-medium mb-2">Block Header</h4>
                                <div className="text-xs space-y-1">
                                    <div>Slot: {client.blockHeader.slot.toString()}</div>
                                    <div>Proposer Index: {client.blockHeader.proposerIndex.toString()}</div>
                                    <div>Block Root: {client.blockRoot}</div>
                                    <div>Parent Root: {client.blockHeader.parentRoot}</div>
                                    <div>State Root: {client.blockHeader.stateRoot}</div>
                                    <div>Body Root: {client.blockHeader.bodyRoot}</div>
                                </div>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
}
