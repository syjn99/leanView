import { useQuery } from '@connectrpc/connect-query';
import { Clock, Database, Globe, Hash, Layers } from 'lucide-react';
import type { FC } from 'react';

import { getAllClientsHeads } from '../gen/proto/api/v1/monitoring-MonitoringService_connectquery';
import {
    Accordion,
    AccordionContent,
    AccordionItem,
    AccordionTrigger,
} from './ui/accordion';
import { Badge } from './ui/badge';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from './ui/card';

export const ClientMonitoringDashboard: FC = () => {
    const { data, isLoading, error } = useQuery(
        getAllClientsHeads,
        {},
        { refetchInterval: 4000 },
    );

    if (isLoading) {
        return (
            <div className='container mx-auto p-6'>
                <Card>
                    <CardContent className='p-6'>
                        <div className='flex items-center justify-center space-x-2'>
                            <div className='animate-spin rounded-full h-4 w-4 border-2 border-primary border-t-transparent'></div>
                            <span className='text-muted-foreground'>
                                Loading client data...
                            </span>
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (error) {
        return (
            <div className='container mx-auto p-6'>
                <Card className='border-destructive'>
                    <CardContent className='p-6'>
                        <div className='text-destructive'>
                            Error: {error.message}
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (!data) {
        return (
            <div className='container mx-auto p-6'>
                <Card>
                    <CardContent className='p-6'>
                        <div className='text-muted-foreground'>
                            No data available
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    return (
        <div className='container mx-auto p-6 space-y-6'>
            {/* Header */}
            <div className='flex flex-col space-y-2'>
                <h1 className='text-3xl font-bold tracking-tight'>
                    Client Monitoring Dashboard
                </h1>
                <p className='text-muted-foreground'>
                    Real-time monitoring of PQ Devnet clients
                </p>
            </div>

            {/* Summary Card */}
            <Card>
                <CardHeader>
                    <CardTitle className='flex items-center space-x-2'>
                        <Database className='h-5 w-5' />
                        <span>Network Overview</span>
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className='flex flex-wrap gap-6'>
                        <div className='flex items-center space-x-2'>
                            <span className='text-sm font-medium'>
                                Total Clients:
                            </span>
                            <Badge variant='outline'>{data.totalClients}</Badge>
                        </div>
                        <div className='flex items-center space-x-2'>
                            <span className='text-sm font-medium'>
                                Healthy Clients:
                            </span>
                            <Badge
                                variant={
                                    data.healthyClients === data.totalClients
                                        ? 'default'
                                        : 'secondary'
                                }
                            >
                                {data.healthyClients}
                            </Badge>
                        </div>
                        <div className='flex items-center space-x-2'>
                            <span className='text-sm font-medium'>
                                Last Update:
                            </span>
                            <Badge
                                variant='outline'
                                className='flex items-center space-x-1'
                            >
                                <Clock className='h-3 w-3' />
                                <span>{new Date().toISOString()}</span>
                            </Badge>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Client Cards Grid */}
            <div className='grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6'>
                {data.clientHeads.map((client, index) => (
                    <Card key={client.clientLabel || index} className='h-fit'>
                        <CardHeader className='pb-3'>
                            <div className='flex items-center justify-between'>
                                <CardTitle className='text-lg flex items-center space-x-2'>
                                    <Globe className='h-4 w-4' />
                                    <span>{client.clientLabel}</span>
                                </CardTitle>
                                <Badge
                                    variant={
                                        client.isHealthy
                                            ? 'default'
                                            : 'destructive'
                                    }
                                >
                                    {client.isHealthy ? 'Healthy' : 'Unhealthy'}
                                </Badge>
                            </div>
                            <CardDescription className='space-y-1'>
                                <div className='flex items-center space-x-1 text-xs'>
                                    <Globe className='h-3 w-3' />
                                    <span>{client.endpointUrl}</span>
                                </div>
                                <div className='flex items-center space-x-1 text-xs'>
                                    <Clock className='h-3 w-3' />
                                    <span>
                                        {new Date(
                                            Number(client.lastUpdateMs),
                                        ).toISOString()}
                                    </span>
                                </div>
                            </CardDescription>
                        </CardHeader>

                        {client.blockHeader && (
                            <CardContent className='pt-0'>
                                <Accordion
                                    type='single'
                                    collapsible
                                    className='w-full'
                                >
                                    <AccordionItem
                                        value='block-details'
                                        className='border-0'
                                    >
                                        <AccordionTrigger className='hover:no-underline py-2'>
                                            <div className='flex items-center space-x-2'>
                                                <Layers className='h-4 w-4' />
                                                <span className='font-medium'>
                                                    Block Details
                                                </span>
                                                <Badge
                                                    variant='secondary'
                                                    className='ml-2'
                                                >
                                                    Slot{' '}
                                                    {client.blockHeader.slot.toString()}
                                                </Badge>
                                                {client.blockRoot && (
                                                    <Badge
                                                        variant='outline'
                                                        className='font-mono text-xs'
                                                    >
                                                        {client.blockRoot.slice(
                                                            0,
                                                            8,
                                                        )}
                                                        ...
                                                    </Badge>
                                                )}
                                            </div>
                                        </AccordionTrigger>
                                        <AccordionContent className='pt-2'>
                                            <div className='grid grid-cols-1 gap-3 text-sm'>
                                                <div className='grid grid-cols-2 gap-2'>
                                                    <div>
                                                        <span className='font-medium text-muted-foreground'>
                                                            Slot:
                                                        </span>
                                                        <div className='font-mono text-xs mt-1'>
                                                            {client.blockHeader.slot.toString()}
                                                        </div>
                                                    </div>
                                                    <div>
                                                        <span className='font-medium text-muted-foreground'>
                                                            Proposer:
                                                        </span>
                                                        <div className='font-mono text-xs mt-1'>
                                                            {client.blockHeader.proposerIndex.toString()}
                                                        </div>
                                                    </div>
                                                </div>

                                                <div>
                                                    <span className='font-medium text-muted-foreground flex items-center space-x-1'>
                                                        <Hash className='h-3 w-3' />
                                                        <span>Block Root:</span>
                                                    </span>
                                                    <div className='font-mono text-xs mt-1 break-all bg-muted p-2 rounded'>
                                                        {client.blockRoot}
                                                    </div>
                                                </div>

                                                <div>
                                                    <span className='font-medium text-muted-foreground flex items-center space-x-1'>
                                                        <Hash className='h-3 w-3' />
                                                        <span>
                                                            Parent Root:
                                                        </span>
                                                    </span>
                                                    <div className='font-mono text-xs mt-1 break-all bg-muted p-2 rounded'>
                                                        {
                                                            client.blockHeader
                                                                .parentRoot
                                                        }
                                                    </div>
                                                </div>

                                                <div>
                                                    <span className='font-medium text-muted-foreground flex items-center space-x-1'>
                                                        <Hash className='h-3 w-3' />
                                                        <span>State Root:</span>
                                                    </span>
                                                    <div className='font-mono text-xs mt-1 break-all bg-muted p-2 rounded'>
                                                        {
                                                            client.blockHeader
                                                                .stateRoot
                                                        }
                                                    </div>
                                                </div>

                                                <div>
                                                    <span className='font-medium text-muted-foreground flex items-center space-x-1'>
                                                        <Hash className='h-3 w-3' />
                                                        <span>Body Root:</span>
                                                    </span>
                                                    <div className='font-mono text-xs mt-1 break-all bg-muted p-2 rounded'>
                                                        {
                                                            client.blockHeader
                                                                .bodyRoot
                                                        }
                                                    </div>
                                                </div>
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>
                                </Accordion>
                            </CardContent>
                        )}
                    </Card>
                ))}
            </div>
        </div>
    );
};
