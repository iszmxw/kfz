declare module "@vant/weapp/toast/toast" {
  type ToastMessage = string | number;

  interface ToastOptions {
    show?: boolean;
    type?: string;
    mask?: boolean;
    zIndex?: number;
    position?: string;
    duration?: number;
    selector?: string;
    forbidClick?: boolean;
    loadingType?: string;
    message?: ToastMessage;
    onClose?: () => void;
  }

  interface ToastFunction {
    (toastOptions: ToastOptions | ToastMessage): WechatMiniprogram.Component.TrivialInstance | undefined;
    loading(options: ToastOptions | ToastMessage): WechatMiniprogram.Component.TrivialInstance | undefined;
    success(options: ToastOptions | ToastMessage): WechatMiniprogram.Component.TrivialInstance | undefined;
    fail(options: ToastOptions | ToastMessage): WechatMiniprogram.Component.TrivialInstance | undefined;
    clear(): void;
  }

  const Toast: ToastFunction;
  export default Toast;
}
