import { useState } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from './ui/table';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { ChevronLeft, ChevronRight, Hash } from 'lucide-react';
import { getBlockHeaders } from '../gen/proto/api/v1/block-BlockService_connectquery';
import { GetBlockHeadersRequest_SortOrder } from '../gen/proto/api/v1/block_pb';
import { keepPreviousData } from '@tanstack/react-query';

export const BlockHeadersTable = () => {
    const [page, setPage] = useState(0);
    const limit = 50;

    const { data, isLoading, error } = useQuery(
        getBlockHeaders,
        {
            limit,
            offset: BigInt(page * limit),
            sortOrder: GetBlockHeadersRequest_SortOrder.SLOT_DESC,
        },
        {
            refetchInterval: 10000, // Refresh every 10 seconds
            placeholderData: keepPreviousData,
        },
    );

    if (error) {
        return (
            <Card className='border-destructive'>
                <CardContent className='p-6'>
                    <div className='text-destructive'>
                        Error loading block headers: {error.message}
                    </div>
                </CardContent>
            </Card>
        );
    }

    const truncateHash = (hash: string | undefined) => {
        if (!hash) return '-';
        return `${hash.slice(0, 10)}...${hash.slice(-8)}`;
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle className='flex items-center space-x-2'>
                    <Hash className='h-5 w-5' />
                    <span>Block Headers</span>
                    {data && (
                        <Badge variant='secondary' className='ml-auto'>
                            Total: {data.totalCount.toLocaleString()}
                        </Badge>
                    )}
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className='rounded-md border'>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className='w-[100px]'>
                                    Slot
                                </TableHead>
                                <TableHead className='w-[120px]'>
                                    Proposer Index
                                </TableHead>
                                <TableHead>Block Root</TableHead>
                                <TableHead>Parent Root</TableHead>
                                <TableHead>State Root</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {isLoading ? (
                                <TableRow>
                                    <TableCell
                                        colSpan={5}
                                        className='text-center'
                                    >
                                        <div className='flex items-center justify-center space-x-2 py-4'>
                                            <div className='animate-spin rounded-full h-4 w-4 border-2 border-primary border-t-transparent' />
                                            <span className='text-muted-foreground'>
                                                Loading block headers...
                                            </span>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            ) : data?.headers?.length === 0 ? (
                                <TableRow>
                                    <TableCell
                                        colSpan={5}
                                        className='text-center text-muted-foreground py-8'
                                    >
                                        No block headers found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                data?.headers?.map((item) => (
                                    <TableRow
                                        key={item.header?.slot?.toString()}
                                    >
                                        <TableCell>
                                            <Badge variant='outline'>
                                                {item.header?.slot?.toString() ||
                                                    '0'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className='text-center'>
                                            {item.header?.proposerIndex?.toString() ||
                                                '0'}
                                        </TableCell>
                                        <TableCell className='font-mono text-xs'>
                                            {truncateHash(item.blockRoot)}
                                        </TableCell>
                                        <TableCell className='font-mono text-xs'>
                                            {truncateHash(
                                                item.header?.parentRoot,
                                            )}
                                        </TableCell>
                                        <TableCell className='font-mono text-xs'>
                                            {truncateHash(
                                                item.header?.stateRoot,
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </div>

                {/* Pagination Controls */}
                {data && data.headers && data.headers.length > 0 && (
                    <div className='flex items-center justify-between space-x-2 py-4'>
                        <div className='text-sm text-muted-foreground'>
                            Showing {page * limit + 1} to{' '}
                            {Math.min((page + 1) * limit, data.totalCount)} of{' '}
                            {data.totalCount.toLocaleString()} entries
                        </div>
                        <div className='flex space-x-2'>
                            <Button
                                variant='outline'
                                size='sm'
                                onClick={() =>
                                    setPage((p) => Math.max(0, p - 1))
                                }
                                disabled={page === 0}
                            >
                                <ChevronLeft className='h-4 w-4 mr-1' />
                                Previous
                            </Button>
                            <div className='flex items-center space-x-1'>
                                <span className='text-sm text-muted-foreground'>
                                    Page {page + 1}
                                </span>
                            </div>
                            <Button
                                variant='outline'
                                size='sm'
                                onClick={() => setPage((p) => p + 1)}
                                disabled={!data.hasMore}
                            >
                                Next
                                <ChevronRight className='h-4 w-4 ml-1' />
                            </Button>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};
