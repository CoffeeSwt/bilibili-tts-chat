package config

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// 环境变量配置结构体
type EnvConfig struct {
	IdCode       string // 身份码，从环境变量 ID_CODE 读取
	TTSAccessKey string // TTS访问密钥
	TTSAppID     string // TTS应用ID
}

// 音色信息结构体 - 用于提供完整的音色信息
type VoiceInfo struct {
	ID          string
	Name        string
	Description string
	Gender      string
}

// 全局配置变量
var (
	envConfig *EnvConfig
	envOnce   sync.Once
)

// 非敏感常量配置
const (
	// 直播平台相关配置
	OpenPlatformHttpHost = "https://live-open.biliapi.com" //开放平台 (线上环境)
	AppId                = 1761135457345

	// TTS API 相关配置 - v3 API
	TTSAPIUrl     = "https://openspeech.bytedance.com/api/v3/tts/unidirectional"
	TTSResourceID = "seed-tts-1.0"     // 豆包语音合成模型1.0
	TTSUID        = "default_user"     // 默认用户ID
	TTSNamespace  = "BidirectionalTTS" // TTS命名空间

	// WebSocket TTS 相关配置
	WSEndpoint      = "wss://openspeech.bytedance.com/api/v1/tts/ws_binary"
	DefaultVoice    = "zh_female_kefunvsheng_mars_bigtts" // 默认音色
	DefaultEncoding = "mp3"                               // 默认编码格式
	DefaultCluster  = "volcano_tts"                       // 默认集群
)

// 音色列表（按固定顺序排列）
var voiceList = []VoiceInfo{
	{ID: "zh_female_kefunvsheng_mars_bigtts", Name: "可爱女声", Description: "可爱活泼的女声", Gender: "female"},
	{ID: "ICL_zh_female_qingyingduoduo_cs_tob", Name: "清莹朵朵", Description: "清脆悦耳的女声", Gender: "female"},
	{ID: "ICL_zh_female_guaiqiaokeer_cs_tob", Name: "乖巧可儿", Description: "乖巧可爱的女声", Gender: "female"},
	{ID: "ICL_zh_female_lixingyuanzi_cs_tob", Name: "理性圆子", Description: "理性温和的女声", Gender: "female"},
	{ID: "ICL_zh_female_qingtiantaotao_cs_tob", Name: "清甜桃桃", Description: "清甜可人的女声", Gender: "female"},
	{ID: "ICL_zh_female_qingxixiaoxue_cs_tob", Name: "清晰小雪", Description: "清晰明亮的女声", Gender: "female"},
	{ID: "ICL_zh_female_qingtianmeimei_cs_tob", Name: "清甜莓莓", Description: "清甜如莓的女声", Gender: "female"},
	{ID: "ICL_zh_female_kailangtingting_cs_tob", Name: "开朗婷婷", Description: "开朗活泼的女声", Gender: "female"},
	{ID: "ICL_zh_female_wenwanshanshan_cs_tob", Name: "温婉珊珊", Description: "温婉柔和的女声", Gender: "female"},
	{ID: "ICL_zh_female_tianmeixiaoyu_cs_tob", Name: "甜美小雨", Description: "甜美清新的女声", Gender: "female"},
	{ID: "ICL_zh_female_reqingaina_cs_tob", Name: "热情艾娜", Description: "热情洋溢的女声", Gender: "female"},
	{ID: "ICL_zh_female_tianmeixiaoju_cs_tob", Name: "甜美小橘", Description: "甜美活泼的女声", Gender: "female"},
	{ID: "ICL_zh_female_lingdongxinxin_cs_tob", Name: "灵动欣欣", Description: "灵动可爱的女声", Gender: "female"},
	{ID: "ICL_zh_female_nuanxinqianqian_cs_tob", Name: "暖心茜茜", Description: "温暖贴心的女声", Gender: "female"},
	{ID: "ICL_zh_female_ruanmengtuanzi_cs_tob", Name: "软萌团子", Description: "软萌可爱的女声", Gender: "female"},
	{ID: "ICL_zh_female_ruanmengtangtang_cs_tob", Name: "软萌糖糖", Description: "软萌甜美的女声", Gender: "female"},
	{ID: "ICL_zh_female_xiuliqianqian_cs_tob", Name: "秀丽倩倩", Description: "秀丽优雅的女声", Gender: "female"},
	{ID: "ICL_zh_female_kaixinxiaohong_cs_tob", Name: "开心小鸿", Description: "开心快乐的女声", Gender: "female"},
	{ID: "zh_female_maomao_conversation_wvae_bigtts", Name: "文静毛毛", Description: "文静温柔的女声", Gender: "female"},
	{ID: "ICL_zh_female_qiuling_v1_tob", Name: "倾心少女", Description: "青春少女的声音", Gender: "female"},
	{ID: "ICL_zh_female_heainainai_tob", Name: "和蔼奶奶", Description: "和蔼慈祥的老年女声", Gender: "elderly_female"},
	{ID: "ICL_zh_female_linjuayi_tob", Name: "邻居阿姨", Description: "亲切的中年女声", Gender: "female"},
	{ID: "zh_female_wenrouxiaoya_moon_bigtts", Name: "温柔小雅", Description: "温柔优雅的女声", Gender: "female"},
	{ID: "zh_female_peiqi_mars_bigtts", Name: "佩奇猪", Description: "可爱的卡通女声", Gender: "child"},
	{ID: "zh_female_wuzetian_mars_bigtts", Name: "武则天", Description: "威严的古风女声", Gender: "female"},
	{ID: "zh_female_gujie_mars_bigtts", Name: "顾姐", Description: "成熟知性的女声", Gender: "female"},
	{ID: "zh_female_yingtaowanzi_mars_bigtts", Name: "樱桃丸子", Description: "甜美可爱的女声", Gender: "female"},
	{ID: "zh_female_shaoergushi_mars_bigtts", Name: "少儿故事", Description: "适合讲故事的女声", Gender: "female"},
	{ID: "zh_female_qiaopinvsheng_mars_bigtts", Name: "俏皮女声", Description: "俏皮活泼的女声", Gender: "female"},
	{ID: "zh_female_jitangmeimei_mars_bigtts", Name: "鸡汤妹妹", Description: "温暖治愈的女声", Gender: "female"},
	{ID: "zh_female_tiexinnvsheng_mars_bigtts", Name: "贴心女声", Description: "贴心温柔的女声", Gender: "female"},
	{ID: "zh_female_mengyatou_mars_bigtts", Name: "萌丫头", Description: "萌萌哒的女声", Gender: "female"},
	{ID: "zh_female_gufengshaoyu_mars_bigtts", Name: "古风少御", Description: "古风韵味的女声", Gender: "female"},
	{ID: "zh_female_wenroushunv_mars_bigtts", Name: "温柔淑女", Description: "温柔淑雅的女声", Gender: "female"},
	{ID: "ICL_zh_male_qinqiexiaozhuo_cs_tob", Name: "亲切小卓", Description: "亲切温和的男声", Gender: "male"},
	{ID: "ICL_zh_male_qingxinmumu_cs_tob", Name: "清新沐沐", Description: "清新自然的男声", Gender: "male"},
	{ID: "ICL_zh_male_shuanglangxiaoyang_cs_tob", Name: "爽朗小阳", Description: "爽朗阳光的男声", Gender: "male"},
	{ID: "ICL_zh_male_qingxinbobo_cs_tob", Name: "清新波波", Description: "清新活力的男声", Gender: "male"},
	{ID: "ICL_zh_male_chenwenmingzai_cs_tob", Name: "沉稳明仔", Description: "沉稳可靠的男声", Gender: "male"},
	{ID: "ICL_zh_male_yangguangyangyang_cs_tob", Name: "阳光洋洋", Description: "阳光开朗的男声", Gender: "male"},
	{ID: "zh_male_M100_conversation_wvae_bigtts", Name: "悠悠君子", Description: "儒雅的男声", Gender: "male"},
	{ID: "ICL_zh_male_buyan_v1_tob", Name: "醇厚低音", Description: "醇厚磁性的男声", Gender: "male"},
	{ID: "ICL_zh_male_BV144_paoxiaoge_v1_tob", Name: "咆哮小哥", Description: "激情澎湃的男声", Gender: "male"},
	{ID: "zh_male_tiancaitongsheng_mars_bigtts", Name: "天才童声", Description: "聪明可爱的童声", Gender: "child"},
	{ID: "zh_male_sunwukong_mars_bigtts", Name: "猴哥", Description: "孙悟空的声音", Gender: "male"},
	{ID: "zh_male_xionger_mars_bigtts", Name: "熊二", Description: "憨厚可爱的熊二", Gender: "male"},
	{ID: "zh_male_chunhui_mars_bigtts", Name: "广告解说", Description: "专业的广告解说声", Gender: "male"},
	{ID: "zh_male_silang_mars_bigtts", Name: "四郎", Description: "古风男声", Gender: "male"},
	{ID: "zh_male_lanxiaoyang_mars_bigtts", Name: "懒音绵宝", Description: "慵懒可爱的男声", Gender: "male"},
	{ID: "zh_male_dongmanhaimian_mars_bigtts", Name: "亮嗓萌仔", Description: "清亮可爱的男声", Gender: "male"},
	{ID: "zh_male_jieshuonansheng_mars_bigtts", Name: "磁性解说男声", Description: "磁性专业的解说声", Gender: "male"},
	{ID: "ICL_zh_male_neiliancaijun_e991be511569_tob", Name: "内敛才俊", Description: "内敛有才的男声", Gender: "male"},
	{ID: "ICL_zh_male_yangyang_v1_tob", Name: "温暖少年", Description: "温暖阳光的少年声", Gender: "male"},
	{ID: "ICL_zh_male_flc_v1_tob", Name: "儒雅公子", Description: "儒雅温文的男声", Gender: "male"},
	{ID: "zh_male_changtianyi_mars_bigtts", Name: "悬疑解说", Description: "神秘的悬疑解说声", Gender: "male"},
	{ID: "zh_male_ruyaqingnian_mars_bigtts", Name: "儒雅青年", Description: "儒雅的青年男声", Gender: "male"},
	{ID: "zh_male_baqiqingshu_mars_bigtts", Name: "霸气青叔", Description: "霸气成熟的男声", Gender: "male"},
	{ID: "zh_male_qingcang_mars_bigtts", Name: "擎苍", Description: "威严霸气的男声", Gender: "male"},
	{ID: "zh_male_yangguangqingnian_mars_bigtts", Name: "活力小哥", Description: "活力四射的男声", Gender: "male"},
	{ID: "zh_male_fanjuanqingnian_mars_bigtts", Name: "反卷青年", Description: "轻松随性的男声", Gender: "male"},
	{ID: "zh_female_vv_uranus_bigtts", Name: "vivi", Description: "通用场景视频配音女声", Gender: "female"},
	{ID: "zh_male_dayi_saturn_bigtts", Name: "大壹", Description: "专业视频配音男声", Gender: "male"},
	{ID: "zh_female_mizai_saturn_bigtts", Name: "黑猫侦探社咪仔", Description: "神秘可爱的侦探女声", Gender: "female"},
	{ID: "zh_female_jitangnv_saturn_bigtts", Name: "鸡汤女", Description: "温暖治愈的鸡汤女声", Gender: "female"},
	{ID: "zh_female_meilinvyou_saturn_bigtts", Name: "魅力女友", Description: "魅力十足的女友声音", Gender: "female"},
	{ID: "zh_female_santongyongns_saturn_bigtts", Name: "流畅女声", Description: "流畅自然的女声", Gender: "female"},
	{ID: "zh_male_ruyayichen_saturn_bigtts", Name: "儒雅逸辰", Description: "儒雅风度的角色扮演男声", Gender: "male"},
	{ID: "saturn_zh_female_keainvsheng_tob", Name: "可爱女生", Description: "青春可爱的角色扮演女声", Gender: "female"},
	{ID: "saturn_zh_female_tiaopigongzhu_tob", Name: "调皮公主", Description: "调皮活泼的公主角色声音", Gender: "female"},
	{ID: "saturn_zh_male_shuanglangshaonian_tob", Name: "爽朗少年", Description: "爽朗阳光的少年角色声音", Gender: "male"},
	{ID: "saturn_zh_male_tiancaitongzhuo_tob", Name: "天才同桌", Description: "聪明机智的天才同桌声音", Gender: "male"},
	{ID: "zh_male_lengkugege_emo_v2_mars_bigtts", Name: "冷酷哥哥", Description: "冷酷多情感男声", Gender: "male"},
	{ID: "ICL_zh_female_bingruoshaonv_tob", Name: "甜心小美", Description: "甜美多情感女声", Gender: "female"},
	{ID: "zh_male_qingcangdianxia_emo_v2_mars_bigtts", Name: "擎苍殿下", Description: "威严多情感男声", Gender: "male"},
	{ID: "zh_female_bingruoshaonv_emo_v2_mars_bigtts", Name: "病弱少女", Description: "柔弱多情感女声", Gender: "female"},
	{ID: "zh_male_qingcang_emo_v2_mars_bigtts", Name: "擎苍", Description: "霸气多情感男声", Gender: "male"},
	{ID: "zh_female_kefunvsheng_emo_v2_mars_bigtts", Name: "可爱女声", Description: "可爱多情感女声", Gender: "female"},
	{ID: "zh_male_ruyaqingnian_emo_v2_mars_bigtts", Name: "儒雅青年", Description: "儒雅多情感男声", Gender: "male"},
	{ID: "zh_female_gujie_emo_v2_mars_bigtts", Name: "顾姐", Description: "成熟多情感女声", Gender: "female"},
	{ID: "zh_male_yangguangqingnian_emo_v2_mars_bigtts", Name: "活力小哥", Description: "活力多情感男声", Gender: "male"},
	{ID: "zh_female_wenrouxiaoya_emo_v2_moon_bigtts", Name: "温柔小雅", Description: "温柔多情感女声", Gender: "female"},
	{ID: "zh_male_shaonianzixin_emo_v2_moon_bigtts", Name: "少年梓辛", Description: "青春多情感男声", Gender: "male"},
	{ID: "ICL_zh_female_wenrouwenya_tob", Name: "温柔文雅", Description: "通用场景温柔女声", Gender: "female"},
	{ID: "zh_male_hupunan_mars_bigtts", Name: "沪普男", Description: "上海普通话口音男声", Gender: "male"},
	{ID: "zh_female_yueyunv_mars_bigtts", Name: "粤语小溏", Description: "粤语口音女声", Gender: "female"},
	{ID: "zh_male_lubanqihao_mars_bigtts", Name: "鲁班七号", Description: "游戏角色口音男声", Gender: "male"},
	{ID: "zh_female_yangmi_mars_bigtts", Name: "林潇", Description: "明星风格女声", Gender: "female"},
	{ID: "zh_female_linzhiling_mars_bigtts", Name: "玲玲姐姐", Description: "港台风格女声", Gender: "female"},
	{ID: "zh_female_jiyejizi2_mars_bigtts", Name: "春日部姐姐", Description: "动漫风格女声", Gender: "female"},
	{ID: "zh_male_tangseng_mars_bigtts", Name: "唐僧", Description: "古典角色男声", Gender: "male"},
	{ID: "zh_male_zhuangzhou_mars_bigtts", Name: "庄周", Description: "古风哲学家男声", Gender: "male"},
	{ID: "zh_male_zhubajie_mars_bigtts", Name: "猪八戒", Description: "憨厚角色男声", Gender: "male"},
	{ID: "zh_female_ganmaodianyin_mars_bigtts", Name: "感冒电音姐姐", Description: "特殊音效女声", Gender: "female"},
	{ID: "zh_female_naying_mars_bigtts", Name: "直率英子", Description: "直率个性女声", Gender: "female"},
	{ID: "zh_female_leidian_mars_bigtts", Name: "女雷神", Description: "威严神话女声", Gender: "female"},
	{ID: "zh_male_yuzhouzixuan_moon_bigtts", Name: "豫州子轩", Description: "河南口音男声", Gender: "male"},
	{ID: "zh_female_daimengchuanmei_moon_bigtts", Name: "呆萌川妹", Description: "四川口音女声", Gender: "female"},
	{ID: "zh_male_guangxiyuanzhou_moon_bigtts", Name: "广西远舟", Description: "广西口音男声", Gender: "male"},
	{ID: "zh_male_zhoujielun_emo_v2_mars_bigtts", Name: "双节棍小哥", Description: "说唱风格男声", Gender: "male"},
	{ID: "zh_female_wanwanxiaohe_moon_bigtts", Name: "湾湾小何", Description: "台湾口音女声", Gender: "female"},
	{ID: "zh_female_wanqudashu_moon_bigtts", Name: "湾区大叔", Description: "湾区口音女声", Gender: "female"},
	{ID: "zh_male_guozhoudege_moon_bigtts", Name: "广州德哥", Description: "广州口音男声", Gender: "male"},
	{ID: "zh_male_haoyuxiaoge_moon_bigtts", Name: "浩宇小哥", Description: "北方口音男声", Gender: "male"},
	{ID: "zh_male_beijingxiaoye_moon_bigtts", Name: "北京小爷", Description: "北京口音男声", Gender: "male"},
	{ID: "zh_male_jingqiangkanye_moon_bigtts", Name: "京腔侃爷", Description: "京腔说话男声", Gender: "male"},
	{ID: "zh_female_meituojieer_moon_bigtts", Name: "妹坨洁儿", Description: "东北口音女声", Gender: "female"},
	{ID: "ICL_zh_female_chunzhenshaonv_e588402fb8ad_tob", Name: "纯真少女", Description: "纯真可爱的少女声音", Gender: "female"},
	{ID: "ICL_zh_male_xiaonaigou_edf58cf28b8b_tob", Name: "奶气小生", Description: "奶气可爱的男生声音", Gender: "male"},
	{ID: "ICL_zh_female_jinglingxiangdao_1beb294a9e3e_tob", Name: "精灵向导", Description: "神秘精灵女声", Gender: "female"},
	{ID: "ICL_zh_male_menyoupingxiaoge_ffed9fc2fee7_tob", Name: "闷油瓶小哥", Description: "沉默寡言的男声", Gender: "male"},
	{ID: "ICL_zh_male_anrenqinzhu_cd62e63dcdab_tob", Name: "黯刃秦主", Description: "威严古风男声", Gender: "male"},
	{ID: "ICL_zh_male_badaozongcai_v1_tob", Name: "霸道总裁", Description: "霸道强势的总裁声音", Gender: "male"},
	{ID: "ICL_zh_female_ganli_v1_tob", Name: "妩媚可人", Description: "妩媚动人的女声", Gender: "female"},
	{ID: "ICL_zh_female_xiangliangya_v1_tob", Name: "邪魅御姐", Description: "邪魅成熟的御姐声音", Gender: "female"},
	{ID: "ICL_zh_male_ms_tob", Name: "嚣张小哥", Description: "嚣张跋扈的男声", Gender: "male"},
	{ID: "ICL_zh_male_you_tob", Name: "油腻大叔", Description: "油腻中年男声", Gender: "male"},
	{ID: "ICL_zh_male_guaogongzi_v1_tob", Name: "孤傲公子", Description: "孤傲不群的公子声音", Gender: "male"},
	{ID: "ICL_zh_male_huzi_v1_tob", Name: "胡子叔叔", Description: "成熟稳重的叔叔声音", Gender: "male"},
	{ID: "ICL_zh_female_luoqing_v1_tob", Name: "性感魅惑", Description: "性感魅惑的女声", Gender: "female"},
	{ID: "ICL_zh_male_bingruogongzi_tob", Name: "病弱公子", Description: "病弱柔美的公子声音", Gender: "male"},
	{ID: "ICL_zh_female_bingjiao3_tob", Name: "邪魅女王", Description: "邪魅强势的女王声音", Gender: "female"},
	{ID: "ICL_zh_male_aomanqingnian_tob", Name: "傲慢青年", Description: "傲慢自大的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_cujingnansheng_tob", Name: "醋精男生", Description: "爱吃醋的男生声音", Gender: "male"},
	{ID: "ICL_zh_male_shuanglangshaonian_tob", Name: "爽朗少年", Description: "爽朗开朗的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_sajiaonanyou_tob", Name: "撒娇男友", Description: "撒娇可爱的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_wenrounanyou_tob", Name: "温柔男友", Description: "温柔体贴的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_wenshunshaonian_tob", Name: "温顺少年", Description: "温顺乖巧的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_naigounanyou_tob", Name: "粘人男友", Description: "粘人可爱的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_sajiaonansheng_tob", Name: "撒娇男生", Description: "撒娇卖萌的男生声音", Gender: "male"},
	{ID: "ICL_zh_male_huoponanyou_tob", Name: "活泼男友", Description: "活泼开朗的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_tianxinanyou_tob", Name: "甜系男友", Description: "甜美温柔的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_huoliqingnian_tob", Name: "活力青年", Description: "充满活力的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_kailangqingnian_tob", Name: "开朗青年", Description: "开朗乐观的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_lengmoxiongzhang_tob", Name: "冷漠兄长", Description: "冷漠疏离的兄长声音", Gender: "male"},
	{ID: "ICL_zh_male_tiancaitongzhuo_tob", Name: "天才同桌", Description: "聪明天才的同桌声音", Gender: "male"},
	{ID: "ICL_zh_male_pianpiangongzi_tob", Name: "翩翩公子", Description: "风度翩翩的公子声音", Gender: "male"},
	{ID: "ICL_zh_male_mengdongqingnian_tob", Name: "懵懂青年", Description: "懵懂青涩的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_lenglianxiongzhang_tob", Name: "冷脸兄长", Description: "冷脸严肃的兄长声音", Gender: "male"},
	{ID: "ICL_zh_male_bingjiaoshaonian_tob", Name: "病娇少年", Description: "病娇性格的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_bingjiaonanyou_tob", Name: "病娇男友", Description: "病娇占有欲强的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_bingruoshaonian_tob", Name: "病弱少年", Description: "病弱柔美的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_yiqishaonian_tob", Name: "意气少年", Description: "意气风发的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_ganjingshaonian_tob", Name: "干净少年", Description: "干净纯真的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_lengmonanyou_tob", Name: "冷漠男友", Description: "冷漠疏离的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_jingyingqingnian_tob", Name: "精英青年", Description: "精英才俊的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_rexueshaonian_tob", Name: "热血少年", Description: "热血激情的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_qingshuangshaonian_tob", Name: "清爽少年", Description: "清爽干净的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_zhongerqingnian_tob", Name: "中二青年", Description: "中二病的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_lingyunqingnian_tob", Name: "凌云青年", Description: "志向高远的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_zifuqingnian_tob", Name: "自负青年", Description: "自负傲慢的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_bujiqingnian_tob", Name: "不羁青年", Description: "不羁放荡的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_ruyajunzi_tob", Name: "儒雅君子", Description: "儒雅温文的君子声音", Gender: "male"},
	{ID: "ICL_zh_male_diyinchenyu_tob", Name: "低音沉郁", Description: "低沉忧郁的男声", Gender: "male"},
	{ID: "ICL_zh_male_lenglianxueba_tob", Name: "冷脸学霸", Description: "冷脸高智商的学霸声音", Gender: "male"},
	{ID: "ICL_zh_male_ruyazongcai_tob", Name: "儒雅总裁", Description: "儒雅风度的总裁声音", Gender: "male"},
	{ID: "ICL_zh_male_shenchenzongcai_tob", Name: "深沉总裁", Description: "深沉内敛的总裁声音", Gender: "male"},
	{ID: "ICL_zh_male_xiaohouye_tob", Name: "小侯爷", Description: "贵族少爷的声音", Gender: "male"},
	{ID: "ICL_zh_male_gugaogongzi_tob", Name: "孤高公子", Description: "孤高清冷的公子声音", Gender: "male"},
	{ID: "ICL_zh_male_zhangjianjunzi_tob", Name: "仗剑君子", Description: "仗剑行侠的君子声音", Gender: "male"},
	{ID: "ICL_zh_male_wenrunxuezhe_tob", Name: "温润学者", Description: "温润如玉的学者声音", Gender: "male"},
	{ID: "ICL_zh_male_qinqieqingnian_tob", Name: "亲切青年", Description: "亲切和善的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_wenrouxuezhang_tob", Name: "温柔学长", Description: "温柔体贴的学长声音", Gender: "male"},
	{ID: "ICL_zh_male_gaolengzongcai_tob", Name: "高冷总裁", Description: "高冷霸道的总裁声音", Gender: "male"},
	{ID: "ICL_zh_male_lengjungaozhi_tob", Name: "冷峻高智", Description: "冷峻高智商的男声", Gender: "male"},
	{ID: "ICL_zh_male_chanruoshaoye_tob", Name: "孱弱少爷", Description: "孱弱体弱的少爷声音", Gender: "male"},
	{ID: "ICL_zh_male_zixinqingnian_tob", Name: "自信青年", Description: "自信满满的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_qingseqingnian_tob", Name: "青涩青年", Description: "青涩纯真的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_xuebatongzhuo_tob", Name: "学霸同桌", Description: "学霸级别的同桌声音", Gender: "male"},
	{ID: "ICL_zh_male_lengaozongcai_tob", Name: "冷傲总裁", Description: "冷傲不群的总裁声音", Gender: "male"},
	{ID: "ICL_zh_male_yuanqishaonian_tob", Name: "元气少年", Description: "元气满满的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_satuoqingnian_tob", Name: "洒脱青年", Description: "洒脱不羁的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_zhishuaiqingnian_tob", Name: "直率青年", Description: "直率坦诚的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_siwenqingnian_tob", Name: "斯文青年", Description: "斯文儒雅的青年声音", Gender: "male"},
	{ID: "ICL_zh_male_junyigongzi_tob", Name: "俊逸公子", Description: "俊逸潇洒的公子声音", Gender: "male"},
	{ID: "ICL_zh_male_zhangjianxiake_tob", Name: "仗剑侠客", Description: "仗剑江湖的侠客声音", Gender: "male"},
	{ID: "ICL_zh_male_jijiaozhineng_tob", Name: "机甲智能", Description: "机械智能的男声", Gender: "male"},
	{ID: "zh_male_naiqimengwa_mars_bigtts", Name: "奶气萌娃", Description: "奶气可爱的萌娃声音", Gender: "child"},
	{ID: "zh_female_popo_mars_bigtts", Name: "婆婆", Description: "慈祥的老奶奶声音", Gender: "elderly_female"},
	{ID: "zh_female_gaolengyujie_moon_bigtts", Name: "高冷御姐", Description: "高冷成熟的御姐声音", Gender: "female"},
	{ID: "zh_male_aojiaobazong_moon_bigtts", Name: "傲娇霸总", Description: "傲娇霸道的总裁声音", Gender: "male"},
	{ID: "zh_female_meilinvyou_moon_bigtts", Name: "魅力女友", Description: "魅力十足的女友声音", Gender: "female"},
	{ID: "zh_male_shenyeboke_moon_bigtts", Name: "深夜播客", Description: "深夜电台主播声音", Gender: "male"},
	{ID: "zh_female_sajiaonvyou_moon_bigtts", Name: "柔美女友", Description: "柔美温柔的女友声音", Gender: "female"},
	{ID: "zh_female_yuanqinvyou_moon_bigtts", Name: "撒娇学妹", Description: "撒娇可爱的学妹声音", Gender: "female"},
	{ID: "ICL_zh_female_huoponvhai_tob", Name: "活泼女孩", Description: "活泼开朗的女孩声音", Gender: "female"},
	{ID: "zh_male_dongfanghaoran_moon_bigtts", Name: "东方浩然", Description: "正气凛然的男声", Gender: "male"},
	{ID: "ICL_zh_male_lvchaxiaoge_tob", Name: "绿茶小哥", Description: "绿茶男的声音", Gender: "male"},
	{ID: "ICL_zh_female_jiaoruoluoli_tob", Name: "娇弱萝莉", Description: "娇弱可爱的萝莉声音", Gender: "female"},
	{ID: "ICL_zh_male_lengdanshuli_tob", Name: "冷淡疏离", Description: "冷淡疏离的男声", Gender: "male"},
	{ID: "ICL_zh_male_hanhoudunshi_tob", Name: "憨厚敦实", Description: "憨厚老实的男声", Gender: "male"},
	{ID: "ICL_zh_female_huopodiaoman_tob", Name: "活泼刁蛮", Description: "活泼刁蛮的女声", Gender: "female"},
	{ID: "ICL_zh_male_guzhibingjiao_tob", Name: "固执病娇", Description: "固执病娇的男声", Gender: "male"},
	{ID: "ICL_zh_male_sajiaonianren_tob", Name: "撒娇粘人", Description: "撒娇粘人的男声", Gender: "male"},
	{ID: "ICL_zh_female_aomanjiaosheng_tob", Name: "傲慢娇声", Description: "傲慢娇气的女声", Gender: "female"},
	{ID: "ICL_zh_male_xiaosasuixing_tob", Name: "潇洒随性", Description: "潇洒随性的男声", Gender: "male"},
	{ID: "ICL_zh_male_guiyishenmi_tob", Name: "诡异神秘", Description: "诡异神秘的男声", Gender: "male"},
	{ID: "ICL_zh_male_ruyacaijun_tob", Name: "儒雅才俊", Description: "儒雅有才的男声", Gender: "male"},
	{ID: "ICL_zh_male_zhengzhiqingnian_tob", Name: "正直青年", Description: "正直善良的青年声音", Gender: "male"},
	{ID: "ICL_zh_female_jiaohannvwang_tob", Name: "娇憨女王", Description: "娇憨可爱的女王声音", Gender: "female"},
	{ID: "ICL_zh_female_bingjiaomengmei_tob", Name: "病娇萌妹", Description: "病娇可爱的萌妹声音", Gender: "female"},
	{ID: "ICL_zh_male_qingsenaigou_tob", Name: "青涩小生", Description: "青涩纯真的小生声音", Gender: "male"},
	{ID: "ICL_zh_male_chunzhenxuedi_tob", Name: "纯真学弟", Description: "纯真可爱的学弟声音", Gender: "male"},
	{ID: "ICL_zh_male_youroubangzhu_tob", Name: "优柔帮主", Description: "优柔寡断的帮主声音", Gender: "male"},
	{ID: "ICL_zh_male_yourougongzi_tob", Name: "优柔公子", Description: "优柔寡断的公子声音", Gender: "male"},
	{ID: "ICL_zh_female_tiaopigongzhu_tob", Name: "调皮公主", Description: "调皮可爱的公主声音", Gender: "female"},
	{ID: "ICL_zh_male_tiexinnanyou_tob", Name: "贴心男友", Description: "贴心体贴的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_shaonianjiangjun_tob", Name: "少年将军", Description: "年轻的将军声音", Gender: "male"},
	{ID: "ICL_zh_male_bingjiaogege_tob", Name: "病娇哥哥", Description: "病娇占有欲强的哥哥声音", Gender: "male"},
	{ID: "ICL_zh_male_xuebanantongzhuo_tob", Name: "学霸男同桌", Description: "学霸级别的男同桌声音", Gender: "male"},
	{ID: "ICL_zh_male_youmoshushu_tob", Name: "幽默叔叔", Description: "幽默风趣的叔叔声音", Gender: "male"},
	{ID: "ICL_zh_female_jiaxiaozi_tob", Name: "假小子", Description: "假小子性格的女声", Gender: "female"},
	{ID: "ICL_zh_male_wenrounantongzhuo_tob", Name: "温柔男同桌", Description: "温柔体贴的男同桌声音", Gender: "male"},
	{ID: "ICL_zh_male_youmodaye_tob", Name: "幽默大爷", Description: "幽默风趣的大爷声音", Gender: "male"},
	{ID: "ICL_zh_male_asmryexiu_tob", Name: "枕边低语", Description: "温柔的枕边低语声音", Gender: "male"},
	{ID: "ICL_zh_male_shenmifashi_tob", Name: "神秘法师", Description: "神秘莫测的法师声音", Gender: "male"},
	{ID: "zh_female_jiaochuan_mars_bigtts", Name: "娇喘女声", Description: "特殊音效女声", Gender: "female"},
	{ID: "zh_male_livelybro_mars_bigtts", Name: "开朗弟弟", Description: "开朗活泼的弟弟声音", Gender: "male"},
	{ID: "zh_female_flattery_mars_bigtts", Name: "谄媚女声", Description: "谄媚讨好的女声", Gender: "female"},
	{ID: "ICL_zh_male_lengjunshangsi_tob", Name: "冷峻上司", Description: "冷峻严肃的上司声音", Gender: "male"},
	{ID: "ICL_zh_male_cujingnanyou_tob", Name: "醋精男友", Description: "爱吃醋的男友声音", Gender: "male"},
	{ID: "ICL_zh_male_fengfashaonian_tob", Name: "风发少年", Description: "意气风发的少年声音", Gender: "male"},
	{ID: "ICL_zh_male_cixingnansang_tob", Name: "磁性男嗓", Description: "磁性迷人的男声", Gender: "male"},
	{ID: "ICL_zh_male_chengshuzongcai_tob", Name: "成熟总裁", Description: "成熟稳重的总裁声音", Gender: "male"},
	{ID: "ICL_zh_male_aojiaojingying_tob", Name: "傲娇精英", Description: "傲娇的精英男声", Gender: "male"},
	{ID: "ICL_zh_male_aojiaogongzi_tob", Name: "傲娇公子", Description: "傲娇的公子声音", Gender: "male"},
	{ID: "ICL_zh_male_badaoshaoye_tob", Name: "霸道少爷", Description: "霸道任性的少爷声音", Gender: "male"},
	{ID: "ICL_zh_male_fuheigongzi_tob", Name: "腹黑公子", Description: "腹黑狡猾的公子声音", Gender: "male"},
	{ID: "ICL_zh_female_nuanxinxuejie_tob", Name: "暖心学姐", Description: "暖心温柔的学姐声音", Gender: "female"},
	{ID: "ICL_zh_female_keainvsheng_tob", Name: "可爱女生", Description: "可爱活泼的女生声音", Gender: "female"},
	{ID: "ICL_zh_female_chengshujiejie_tob", Name: "成熟姐姐", Description: "成熟知性的姐姐声音", Gender: "female"},
	{ID: "ICL_zh_female_bingjiaojiejie_tob", Name: "病娇姐姐", Description: "病娇占有欲强的姐姐声音", Gender: "female"},
	{ID: "ICL_zh_female_wumeiyujie_tob", Name: "妩媚御姐", Description: "妩媚成熟的御姐声音", Gender: "female"},
	{ID: "ICL_zh_female_aojiaonvyou_tob", Name: "傲娇女友", Description: "傲娇可爱的女友声音", Gender: "female"},
	{ID: "ICL_zh_female_tiexinnvyou_tob", Name: "贴心女友", Description: "贴心体贴的女友声音", Gender: "female"},
	{ID: "ICL_zh_female_xingganyujie_tob", Name: "性感御姐", Description: "性感迷人的御姐声音", Gender: "female"},
	{ID: "ICL_zh_male_bingjiaodidi_tob", Name: "病娇弟弟", Description: "病娇可爱的弟弟声音", Gender: "male"},
	{ID: "ICL_zh_male_aomanshaoye_tob", Name: "傲慢少爷", Description: "傲慢自大的少爷声音", Gender: "male"},
	{ID: "ICL_zh_male_aiqilingren_tob", Name: "傲气凌人", Description: "傲气凌人的男声", Gender: "male"},
	{ID: "ICL_zh_male_bingjiaobailian_tob", Name: "病娇白莲", Description: "病娇白莲花的男声", Gender: "male"},
}

// 初始化随机种子和环境变量
func init() {
	rand.Seed(time.Now().UnixNano())

	// 尝试加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Printf("未找到 .env 文件或加载失败: %v", err)
		log.Println("将使用系统环境变量")
	}
}

// LoadEnvConfig 加载环境变量配置
func LoadEnvConfig() error {
	var err error
	envOnce.Do(func() {
		envConfig = &EnvConfig{
			IdCode:       getEnvWithDefault("ID_CODE", ""),
			TTSAccessKey: getEnvWithDefault("TTS_ACCESS_KEY", ""),
			TTSAppID:     getEnvWithDefault("TTS_APP_ID", ""),
		}

		// 验证必需的环境变量
		if err = validateEnvConfig(envConfig); err != nil {
			log.Printf("环境变量配置验证失败: %v", err)
		} else {
			log.Println("环境变量配置加载成功")
		}
	})
	return err
}

// getEnvWithDefault 获取环境变量，如果不存在则返回默认值
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateEnvConfig 验证环境变量配置
func validateEnvConfig(config *EnvConfig) error {
	var missingVars []string

	if config.IdCode == "" {
		missingVars = append(missingVars, "ID_CODE")
	}
	if config.TTSAccessKey == "" {
		missingVars = append(missingVars, "TTS_ACCESS_KEY")
	}
	if config.TTSAppID == "" {
		missingVars = append(missingVars, "TTS_APP_ID")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("缺少必需的环境变量: %v", missingVars)
	}

	return nil
}

// GetEnvConfig 获取环境变量配置实例
func GetEnvConfig() *EnvConfig {
	if envConfig == nil {
		if err := LoadEnvConfig(); err != nil {
			log.Printf("加载环境变量配置失败: %v", err)
			log.Println("请确保设置了所有必需的环境变量")
		}
	}
	return envConfig
}

// GetIdCode 获取主播身份码
func GetIdCode() string {
	envConfig := GetEnvConfig()
	if envConfig != nil && envConfig.IdCode != "" {
		return envConfig.IdCode
	}
	log.Println("警告: 未找到主播身份码配置，请设置 ID_CODE 环境变量")
	return ""
}

// GetTTSAccessKey 获取TTS Access Key
func GetTTSAccessKey() string {
	envConfig := GetEnvConfig()
	if envConfig != nil && envConfig.TTSAccessKey != "" {
		return envConfig.TTSAccessKey
	}
	log.Println("警告: 未找到TTS Access Key配置，请设置 TTS_ACCESS_KEY 环境变量")
	return ""
}

// GetTTSAppID 获取TTS App ID
func GetTTSAppID() string {
	envConfig := GetEnvConfig()
	if envConfig != nil && envConfig.TTSAppID != "" {
		return envConfig.TTSAppID
	}
	log.Println("警告: 未找到TTS App ID配置，请设置 TTS_APP_ID 环境变量")
	return ""
}

// GetAvailableVoices 获取所有可用的音色ID列表（内置）
func GetAvailableVoices() []string {
	voices := make([]string, 0, len(voiceList))
	for _, voice := range voiceList {
		voices = append(voices, voice.ID)
	}
	return voices
}

// GetRandomVoiceID 随机获取一个音色ID
func GetRandomVoiceID() string {
	voiceIDs := GetAvailableVoices()
	if len(voiceIDs) == 0 {
		return DefaultVoice
	}
	index := rand.Intn(len(voiceIDs))
	return voiceIDs[index]
}

// GetRandomVoiceInfo 随机获取一个音色的完整信息
func GetRandomVoiceInfo() VoiceInfo {
	voiceID := GetRandomVoiceID()
	return GetVoiceInfoByID(voiceID)
}

// GetVoiceInfoByID 根据ID获取音色的完整信息
func GetVoiceInfoByID(id string) VoiceInfo {
	for _, voice := range voiceList {
		if voice.ID == id {
			return voice
		}
	}
	// 如果映射表中没有，返回一个默认的音色信息
	return VoiceInfo{
		ID:          "zh_female_kefunvsheng_mars_bigtts",
		Name:        "可爱女声",
		Description: "可爱活泼的女声",
		Gender:      "female",
	}
}

// GetVoiceInfoByName 根据名称获取音色的完整信息
func GetVoiceInfoByName(name string) VoiceInfo {
	for _, voice := range voiceList {
		if voice.Name == name {
			return voice
		}
	}
	// 如果映射表中没有，返回一个默认的音色信息
	return VoiceInfo{
		ID:          "zh_female_kefunvsheng_mars_bigtts",
		Name:        "可爱女声",
		Description: "可爱活泼的女声",
		Gender:      "female",
	}
}

// GetVoiceByIndex 根据编号获取音色信息（编号从1开始）
func GetVoiceByIndex(index int) VoiceInfo {
	if index < 1 || index > len(voiceList) {
		// 返回空的VoiceInfo表示未找到
		return VoiceInfo{}
	}
	return voiceList[index-1] // 数组索引从0开始，编号从1开始
}

// GetVoiceIndexByID 根据音色ID获取编号（编号从1开始）
func GetVoiceIndexByID(id string) int {
	for i, voice := range voiceList {
		if voice.ID == id {
			return i + 1 // 数组索引从0开始，编号从1开始
		}
	}
	return 0 // 未找到返回0
}
