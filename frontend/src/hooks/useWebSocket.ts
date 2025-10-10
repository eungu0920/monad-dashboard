import { useEffect, useState, useRef } from 'react';
import { io, Socket } from 'socket.io-client';
import type { MonadMetrics } from '../types';

export function useWebSocket() {
  const [metrics, setMetrics] = useState<MonadMetrics | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connect = () => {
    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/ws`;

      console.log(`Attempting to connect to WebSocket: ${wsUrl}`);

      const socket = new WebSocket(wsUrl);
      socketRef.current = socket;

      socket.onopen = () => {
        console.log('WebSocket connected');
        setConnected(true);
        setError(null);
      };

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setMetrics(data);
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
          setError('Failed to parse server data');
        }
      };

      socket.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setConnected(false);
        socketRef.current = null;

        // Attempt to reconnect after 3 seconds
        if (!event.wasClean) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log('Attempting to reconnect...');
            connect();
          }, 3000);
        }
      };

      socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('Connection error');
        setConnected(false);
      };

    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to connect to server');
    }
  };

  const disconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (socketRef.current) {
      socketRef.current.close(1000, 'User disconnected');
      socketRef.current = null;
    }

    setConnected(false);
    setMetrics(null);
  };

  useEffect(() => {
    connect();

    return () => {
      disconnect();
    };
  }, []);

  return {
    metrics,
    connected,
    error,
    reconnect: connect,
  };
}