import { ReactElement, useEffect, useMemo, useState } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router';
import { Layout, Menu, Button, Typography, Dropdown, Avatar, Space, Message } from '@arco-design/web-react';
import {
  IconBook,
  IconCheckCircle,
  IconDashboard,
  IconHistory,
  IconSafe,
  IconSettings,
  IconStorage,
  IconUpload,
  IconUser
} from '@arco-design/web-react/icon';
import { clearToken, get, getToken, post, setToken } from './api';
import type { AdminMenu, AdminUser, LoginResponse } from './types';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import ManualReviewPage from './pages/ManualReviewPage';
import ScanLogPage from './pages/ScanLogPage';
import BookPage from './pages/BookPage';
import PriceSnapshotPage from './pages/PriceSnapshotPage';
import ImportPage from './pages/ImportPage';
import RecycleRulePage from './pages/RecycleRulePage';
import SystemPage from './pages/SystemPage';

const { Header, Sider, Content } = Layout;
const MenuItem = Menu.Item;

const iconMap: Record<string, ReactElement> = {
  IconDashboard: <IconDashboard />,
  IconCheckCircle: <IconCheckCircle />,
  IconHistory: <IconHistory />,
  IconBook: <IconBook />,
  IconStorage: <IconStorage />,
  IconUpload: <IconUpload />,
  IconSettings: <IconSettings />,
  IconSafe: <IconSafe />
};

function AppShell() {
  const [user, setUser] = useState<AdminUser | null>(null);
  const [menus, setMenus] = useState<AdminMenu[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!getToken()) {
      setLoading(false);
      return;
    }
    get<{ user: AdminUser; permissions: string[] }>('/auth/me')
      .then((data) => {
        setUser(data.user);
        return get<AdminMenu[]>('/auth/menus');
      })
      .then(setMenus)
      .catch(() => {
        clearToken();
        navigate('/login');
      })
      .finally(() => setLoading(false));
  }, [navigate]);

  const selectedKey = useMemo(() => {
    const path = location.pathname.replace(/^\/admin/, '') || '/dashboard';
    return [path];
  }, [location.pathname]);

  if (loading) return null;
  if (!user && location.pathname !== '/login') return <Navigate to="/login" replace />;

  if (location.pathname === '/login') {
    return (
      <LoginPage
        onLogin={(data: LoginResponse) => {
          setToken(data.token);
          setUser(data.user);
          setMenus(data.menus);
          navigate('/dashboard');
        }}
      />
    );
  }

  const logout = async () => {
    try {
      await post('/auth/logout');
    } finally {
      clearToken();
      setUser(null);
      Message.success('已退出');
      navigate('/login');
    }
  };

  return (
    <Layout className="app-shell">
      <Sider width={232} className="app-sider">
        <div className="brand">旧书回收后台</div>
        <Menu selectedKeys={selectedKey} onClickMenuItem={(key) => navigate(key)}>
          {menus.map((menu) => (
            <MenuItem key={menu.path}>
              {iconMap[menu.icon] || <IconDashboard />}
              {menu.title}
            </MenuItem>
          ))}
        </Menu>
      </Sider>
      <Layout>
        <Header className="app-header">
          <Typography.Title heading={5} style={{ margin: 0 }}>
            运营闭环管理
          </Typography.Title>
          <Dropdown droplist={<Menu><MenuItem key="logout" onClick={logout}>退出登录</MenuItem></Menu>}>
            <Button type="text">
              <Space>
                <Avatar size={28}><IconUser /></Avatar>
                {user?.name || user?.username}
              </Space>
            </Button>
          </Dropdown>
        </Header>
        <Content className="app-content">
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/manual-review" element={<ManualReviewPage />} />
            <Route path="/scan-log" element={<ScanLogPage />} />
            <Route path="/book" element={<BookPage />} />
            <Route path="/price-snapshot" element={<PriceSnapshotPage />} />
            <Route path="/import" element={<ImportPage />} />
            <Route path="/recycle-rule" element={<RecycleRulePage />} />
            <Route path="/system" element={<SystemPage />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  );
}

export default function App() {
  return (
    <BrowserRouter basename="/admin">
      <AppShell />
    </BrowserRouter>
  );
}
