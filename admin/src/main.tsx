import React from 'react';
import ReactDOM from 'react-dom/client';
import { ConfigProvider } from '@arco-design/web-react';
import zhCN from '@arco-design/web-react/es/locale/zh-CN';
import '@arco-design/web-react/dist/css/arco.css';
import './styles.css';
import App from './App';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <ConfigProvider locale={zhCN} size="default">
      <App />
    </ConfigProvider>
  </React.StrictMode>
);
