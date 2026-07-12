import React from 'react';
import ReactDOM from 'react-dom/client';
import { KcPage } from './login/KcPage';
import './styles.css';

const root = ReactDOM.createRoot(document.getElementById('root')!);
root.render(<React.StrictMode><KcPage /></React.StrictMode>);
