import React, { useEffect, useState } from 'react';

import { HubStats as HubStatsType, useGetHubStats } from '@/types';
import * as Accordion from '@radix-ui/react-accordion';
import { ChevronDownIcon } from 'lucide-react'; // Using Lucide Icons for the Chevron

export const HubStatsAccordion: React.FC<{ API_BASE_URL: string }> = ({ API_BASE_URL }) => {
  const [hubStats, setHubStats] = useState<HubStatsType | null>(null);
  const { data, isLoading, error } = useGetHubStats(
    {
      query: {
        enabled: true,
        refetchInterval: 1000,
      },
    }
  );

  useEffect(() => {
    if (data?.data) {
      setHubStats(data.data);
    } else {
      setHubStats(null);
    }
  }, [data]);

  return (
    <Accordion.Root type="single" collapsible className="w-full">
      <Accordion.Item value="hub-stats" className="border border-gray-200 rounded-md mb-4">
        <Accordion.Header>
          <Accordion.Trigger className="flex justify-between items-center w-full p-4 rounded-md cursor-pointer focus:outline-none">
            <span className="text-sm text-gray-400">Websocket Hub Statistics</span>
            <ChevronDownIcon className="h-5 w-5 transition-transform duration-200" />
          </Accordion.Trigger>
        </Accordion.Header>
        <Accordion.Content className="p-4 bg-white rounded-md">
          {hubStats ? (
            <HubStats stats={hubStats} />
          ) : (
            <p>Loading hub statistics...</p>
          )}
        </Accordion.Content>
      </Accordion.Item>
    </Accordion.Root>
  );
};

const HubStats: React.FC<{ stats: HubStatsType }> = ({ stats }) => {
  return (
    <div className="bg-gray-100 p-4 rounded-lg">
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
        <StatItem label="Reviews waiting to be assigned (server-side)" value={stats.pending_reviews_count} />
        <StatItem label="Reviews in progress (client-side)" value={stats.assigned_reviews_count} />
        <StatItem label="Completed Reviews" value={stats.completed_reviews_count} />
      </div>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <StatItem label="Connected Clients" value={stats.connected_clients} />
        <StatItem label="Free Clients" value={stats.free_clients} />
        <StatItem label="Busy Clients" value={stats.busy_clients} />
      </div>
      <div className="mt-4">
        <h3 className="font-semibold mb-2">Assigned Reviews</h3>
        <ul className="list-disc list-inside">
          {Object.entries(stats.assigned_reviews).map(([client, count]) => (
            <li key={client}>Client {client.slice(-6)}: {count}</li>
          ))}
        </ul>
      </div>
      <div className="mt-4">
        <h3 className="font-semibold mb-2">Review Distribution</h3>
        <ul className="list-disc list-inside">
          {Object.entries(stats.review_distribution).map(([reviewCount, clientCount]) => (
            <li key={reviewCount}>{clientCount} client(s) with {reviewCount} review(s)</li>
          ))}
        </ul>
      </div>
    </div>
  );
};

const StatItem: React.FC<{ label: string; value: number }> = ({ label, value }) => (
  <div className="bg-white p-3 rounded shadow">
    <div className="text-sm text-gray-600">{label}</div>
    <div className="text-xl font-semibold">{value}</div>
  </div>
);
