package live

import (
	"sync"
	"testing"
)

func TestStart(t *testing.T) {
	t.Run("douyin-live", func(t *testing.T) {
		handler := NewFastHttpHandler(WithHeader(map[string]string{
			"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36",
			"cookie":     "__ac_nonce=0638733a400869171be51",
		}), WithRequestURI("https://live.douyin.com/456066026839"))
		body, err := handler.FastDo()
		t.Log(body, err)
		Start(body, "wss://webcast5-ws-web-lf.douyin.com/webcast/im/push/v2/?app_name=douyin_web&version_code=180800&webcast_sdk_version=1.0.12&update_version_code=1.0.12&compress=gzip&device_platform=web&cookie_enabled=true&screen_width=2560&screen_height=1440&browser_language=zh-CN&browser_platform=MacIntel&browser_name=Mozilla&browser_version=5.0%20(Macintosh;%20Intel%20Mac%20OS%20X%2010_15_7)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/123.0.0.0%20Safari/537.36&browser_online=true&tz_name=Asia/Shanghai&cursor=d-1_u-1_fh-7353165541137060890_t-1712042298264_r-1&internal_ext=internal_src:dim|wss_push_room_id:7353122286687308582|wss_push_did:7353138815589303820|first_req_ms:1712042298156|fetch_time:1712042298264|seq:1|wss_info:0-1712042298264-0-0|wrds_v:7353165670688757172&host=https://live.douyin.com&aid=6383&live_id=1&did_rule=3&endpoint=live_pc&support_wrds=1&user_unique_id=7353138815589303820&im_path=/webcast/im/fetch/&identity=audience&need_persist_msg_count=15&insert_task_id=&live_reason=&room_id=7353122286687308582&heartbeatDuration=0&signature=WBr1t2tEWnjYrk+l")
		var (
			wg sync.WaitGroup
		)
		wg.Add(1)
		wg.Wait()
	})
}
