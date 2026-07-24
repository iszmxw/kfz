import { initAudioFeedback, setAudioFeedbackForeground } from "./utils/feedback";

App<IAppOption>({
  globalData: {},
  onLaunch() {
    initAudioFeedback();
  },
  onShow() {
    setAudioFeedbackForeground(true);
  },
  onHide() {
    setAudioFeedbackForeground(false);
  }
});
