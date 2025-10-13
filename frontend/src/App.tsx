import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { FiredancerLayout } from './components/FiredancerLayout';
import { Overview, LeaderSchedule, Gossip } from './pages';
import './index.css';

function App() {
  return (
    <BrowserRouter>
      <FiredancerLayout>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/leader-schedule" element={<LeaderSchedule />} />
          <Route path="/gossip" element={<Gossip />} />
        </Routes>
      </FiredancerLayout>
    </BrowserRouter>
  );
}

export default App;
