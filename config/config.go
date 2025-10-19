package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// 音色配置结构体 - 简化版，只包含ID
type Voice struct {
	ID string `yaml:"id"`
}

// YAML配置结构体
type YAMLConfig struct {
	IdCode string  `yaml:"id_code"`
	Voices []Voice `yaml:"voices"`
}

// 环境变量配置结构体 - 简化版，只保留TTS相关配置
type EnvConfig struct {
	TTSAccessKey string
	TTSAppID     string
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
	yamlConfig     *YAMLConfig
	envConfig      *EnvConfig
	configOnce     sync.Once
	envOnce        sync.Once
	configFilePath = "config.yaml"
)

// 非敏感常量配置
const (
	// 直播平台相关配置
	OpenPlatformHttpHost = "https://live-open.biliapi.com" //开放平台 (线上环境)
	AppId                = 1761135457345

	// TTS API 相关配置 - v3 API
	TTSAPIUrl     = "https://openspeech.bytedance.com/api/v3/tts/unidirectional"
	TTSResourceID = "seed-tts-1.0"   // 豆包语音合成模型1.0
	TTSUID        = "default_user"   // 默认用户ID
	TTSNamespace  = "BidirectionalTTS" // TTS命名空间

	// WebSocket TTS 相关配置
	WSEndpoint      = "wss://openspeech.bytedance.com/api/v1/tts/ws_binary"
	DefaultVoice    = "zh_female_kefunvsheng_mars_bigtts" // 默认音色
	DefaultEncoding = "mp3"                               // 默认编码格式
	DefaultCluster  = "volcano_tts"                       // 默认集群
)

// 音色ID到名称的映射表
var voiceNameMap = map[string]VoiceInfo{
	// 女声音色
	"zh_female_kefunvsheng_mars_bigtts":           {ID: "zh_female_kefunvsheng_mars_bigtts", Name: "可爱女声", Description: "可爱活泼的女声", Gender: "female"},
	"ICL_zh_female_qingyingduoduo_cs_tob":         {ID: "ICL_zh_female_qingyingduoduo_cs_tob", Name: "清莹朵朵", Description: "清脆悦耳的女声", Gender: "female"},
	"ICL_zh_female_guaiqiaokeer_cs_tob":           {ID: "ICL_zh_female_guaiqiaokeer_cs_tob", Name: "乖巧可儿", Description: "乖巧可爱的女声", Gender: "female"},
	"ICL_zh_female_lixingyuanzi_cs_tob":           {ID: "ICL_zh_female_lixingyuanzi_cs_tob", Name: "理性圆子", Description: "理性温和的女声", Gender: "female"},
	"ICL_zh_female_qingtiantaotao_cs_tob":         {ID: "ICL_zh_female_qingtiantaotao_cs_tob", Name: "清甜桃桃", Description: "清甜可人的女声", Gender: "female"},
	"ICL_zh_female_qingxixiaoxue_cs_tob":          {ID: "ICL_zh_female_qingxixiaoxue_cs_tob", Name: "清晰小雪", Description: "清晰明亮的女声", Gender: "female"},
	"ICL_zh_female_qingtianmeimei_cs_tob":         {ID: "ICL_zh_female_qingtianmeimei_cs_tob", Name: "清甜莓莓", Description: "清甜如莓的女声", Gender: "female"},
	"ICL_zh_female_kailangtingting_cs_tob":        {ID: "ICL_zh_female_kailangtingting_cs_tob", Name: "开朗婷婷", Description: "开朗活泼的女声", Gender: "female"},
	"ICL_zh_female_wenwanshanshan_cs_tob":         {ID: "ICL_zh_female_wenwanshanshan_cs_tob", Name: "温婉珊珊", Description: "温婉柔和的女声", Gender: "female"},
	"ICL_zh_female_tianmeixiaoyu_cs_tob":          {ID: "ICL_zh_female_tianmeixiaoyu_cs_tob", Name: "甜美小雨", Description: "甜美清新的女声", Gender: "female"},
	"ICL_zh_female_reqingaina_cs_tob":             {ID: "ICL_zh_female_reqingaina_cs_tob", Name: "热情艾娜", Description: "热情洋溢的女声", Gender: "female"},
	"ICL_zh_female_tianmeixiaoju_cs_tob":          {ID: "ICL_zh_female_tianmeixiaoju_cs_tob", Name: "甜美小橘", Description: "甜美活泼的女声", Gender: "female"},
	"ICL_zh_female_lingdongxinxin_cs_tob":         {ID: "ICL_zh_female_lingdongxinxin_cs_tob", Name: "灵动欣欣", Description: "灵动可爱的女声", Gender: "female"},
	"ICL_zh_female_nuanxinqianqian_cs_tob":        {ID: "ICL_zh_female_nuanxinqianqian_cs_tob", Name: "暖心茜茜", Description: "温暖贴心的女声", Gender: "female"},
	"ICL_zh_female_ruanmengtuanzi_cs_tob":         {ID: "ICL_zh_female_ruanmengtuanzi_cs_tob", Name: "软萌团子", Description: "软萌可爱的女声", Gender: "female"},
	"ICL_zh_female_ruanmengtangtang_cs_tob":       {ID: "ICL_zh_female_ruanmengtangtang_cs_tob", Name: "软萌糖糖", Description: "软萌甜美的女声", Gender: "female"},
	"ICL_zh_female_xiuliqianqian_cs_tob":          {ID: "ICL_zh_female_xiuliqianqian_cs_tob", Name: "秀丽倩倩", Description: "秀丽优雅的女声", Gender: "female"},
	"ICL_zh_female_kaixinxiaohong_cs_tob":         {ID: "ICL_zh_female_kaixinxiaohong_cs_tob", Name: "开心小鸿", Description: "开心快乐的女声", Gender: "female"},
	"zh_female_maomao_conversation_wvae_bigtts":   {ID: "zh_female_maomao_conversation_wvae_bigtts", Name: "文静毛毛", Description: "文静温柔的女声", Gender: "female"},
	"ICL_zh_female_qiuling_v1_tob":                {ID: "ICL_zh_female_qiuling_v1_tob", Name: "倾心少女", Description: "青春少女的声音", Gender: "female"},
	"ICL_zh_female_heainainai_tob":                {ID: "ICL_zh_female_heainainai_tob", Name: "和蔼奶奶", Description: "和蔼慈祥的老年女声", Gender: "elderly_female"},
	"ICL_zh_female_linjuayi_tob":                  {ID: "ICL_zh_female_linjuayi_tob", Name: "邻居阿姨", Description: "亲切的中年女声", Gender: "female"},
	"zh_female_wenrouxiaoya_moon_bigtts":          {ID: "zh_female_wenrouxiaoya_moon_bigtts", Name: "温柔小雅", Description: "温柔优雅的女声", Gender: "female"},
	"zh_female_peiqi_mars_bigtts":                 {ID: "zh_female_peiqi_mars_bigtts", Name: "佩奇猪", Description: "可爱的卡通女声", Gender: "child"},
	"zh_female_wuzetian_mars_bigtts":              {ID: "zh_female_wuzetian_mars_bigtts", Name: "武则天", Description: "威严的古风女声", Gender: "female"},
	"zh_female_gujie_mars_bigtts":                 {ID: "zh_female_gujie_mars_bigtts", Name: "顾姐", Description: "成熟知性的女声", Gender: "female"},
	"zh_female_yingtaowanzi_mars_bigtts":          {ID: "zh_female_yingtaowanzi_mars_bigtts", Name: "樱桃丸子", Description: "甜美可爱的女声", Gender: "female"},
	"zh_female_shaoergushi_mars_bigtts":           {ID: "zh_female_shaoergushi_mars_bigtts", Name: "少儿故事", Description: "适合讲故事的女声", Gender: "female"},
	"zh_female_qiaopinvsheng_mars_bigtts":         {ID: "zh_female_qiaopinvsheng_mars_bigtts", Name: "俏皮女声", Description: "俏皮活泼的女声", Gender: "female"},
	"zh_female_jitangmeimei_mars_bigtts":          {ID: "zh_female_jitangmeimei_mars_bigtts", Name: "鸡汤妹妹", Description: "温暖治愈的女声", Gender: "female"},
	"zh_female_tiexinnvsheng_mars_bigtts":         {ID: "zh_female_tiexinnvsheng_mars_bigtts", Name: "贴心女声", Description: "贴心温柔的女声", Gender: "female"},
	"zh_female_mengyatou_mars_bigtts":             {ID: "zh_female_mengyatou_mars_bigtts", Name: "萌丫头", Description: "萌萌哒的女声", Gender: "female"},
	"zh_female_gufengshaoyu_mars_bigtts":          {ID: "zh_female_gufengshaoyu_mars_bigtts", Name: "古风少御", Description: "古风韵味的女声", Gender: "female"},
	"zh_female_wenroushunv_mars_bigtts":           {ID: "zh_female_wenroushunv_mars_bigtts", Name: "温柔淑女", Description: "温柔淑雅的女声", Gender: "female"},

	// 男声音色
	"ICL_zh_male_qinqiexiaozhuo_cs_tob":           {ID: "ICL_zh_male_qinqiexiaozhuo_cs_tob", Name: "亲切小卓", Description: "亲切温和的男声", Gender: "male"},
	"ICL_zh_male_qingxinmumu_cs_tob":              {ID: "ICL_zh_male_qingxinmumu_cs_tob", Name: "清新沐沐", Description: "清新自然的男声", Gender: "male"},
	"ICL_zh_male_shuanglangxiaoyang_cs_tob":       {ID: "ICL_zh_male_shuanglangxiaoyang_cs_tob", Name: "爽朗小阳", Description: "爽朗阳光的男声", Gender: "male"},
	"ICL_zh_male_qingxinbobo_cs_tob":              {ID: "ICL_zh_male_qingxinbobo_cs_tob", Name: "清新波波", Description: "清新活力的男声", Gender: "male"},
	"ICL_zh_male_chenwenmingzai_cs_tob":           {ID: "ICL_zh_male_chenwenmingzai_cs_tob", Name: "沉稳明仔", Description: "沉稳可靠的男声", Gender: "male"},
	"ICL_zh_male_yangguangyangyang_cs_tob":        {ID: "ICL_zh_male_yangguangyangyang_cs_tob", Name: "阳光洋洋", Description: "阳光开朗的男声", Gender: "male"},
	"zh_male_M100_conversation_wvae_bigtts":       {ID: "zh_male_M100_conversation_wvae_bigtts", Name: "悠悠君子", Description: "儒雅的男声", Gender: "male"},
	"ICL_zh_male_buyan_v1_tob":                    {ID: "ICL_zh_male_buyan_v1_tob", Name: "醇厚低音", Description: "醇厚磁性的男声", Gender: "male"},
	"ICL_zh_male_BV144_paoxiaoge_v1_tob":          {ID: "ICL_zh_male_BV144_paoxiaoge_v1_tob", Name: "咆哮小哥", Description: "激情澎湃的男声", Gender: "male"},
	"zh_male_tiancaitongsheng_mars_bigtts":        {ID: "zh_male_tiancaitongsheng_mars_bigtts", Name: "天才童声", Description: "聪明可爱的童声", Gender: "child"},
	"zh_male_sunwukong_mars_bigtts":               {ID: "zh_male_sunwukong_mars_bigtts", Name: "猴哥", Description: "孙悟空的声音", Gender: "male"},
	"zh_male_xionger_mars_bigtts":                 {ID: "zh_male_xionger_mars_bigtts", Name: "熊二", Description: "憨厚可爱的熊二", Gender: "male"},
	"zh_male_chunhui_mars_bigtts":                 {ID: "zh_male_chunhui_mars_bigtts", Name: "广告解说", Description: "专业的广告解说声", Gender: "male"},
	"zh_male_silang_mars_bigtts":                  {ID: "zh_male_silang_mars_bigtts", Name: "四郎", Description: "古风男声", Gender: "male"},
	"zh_male_lanxiaoyang_mars_bigtts":             {ID: "zh_male_lanxiaoyang_mars_bigtts", Name: "懒音绵宝", Description: "慵懒可爱的男声", Gender: "male"},
	"zh_male_dongmanhaimian_mars_bigtts":          {ID: "zh_male_dongmanhaimian_mars_bigtts", Name: "亮嗓萌仔", Description: "清亮可爱的男声", Gender: "male"},
	"zh_male_jieshuonansheng_mars_bigtts":         {ID: "zh_male_jieshuonansheng_mars_bigtts", Name: "磁性解说男声", Description: "磁性专业的解说声", Gender: "male"},
	"ICL_zh_male_neiliancaijun_e991be511569_tob":  {ID: "ICL_zh_male_neiliancaijun_e991be511569_tob", Name: "内敛才俊", Description: "内敛有才的男声", Gender: "male"},
	"ICL_zh_male_yangyang_v1_tob":                 {ID: "ICL_zh_male_yangyang_v1_tob", Name: "温暖少年", Description: "温暖阳光的少年声", Gender: "male"},
	"ICL_zh_male_flc_v1_tob":                      {ID: "ICL_zh_male_flc_v1_tob", Name: "儒雅公子", Description: "儒雅温文的男声", Gender: "male"},
	"zh_male_changtianyi_mars_bigtts":             {ID: "zh_male_changtianyi_mars_bigtts", Name: "悬疑解说", Description: "神秘的悬疑解说声", Gender: "male"},
	"zh_male_ruyaqingnian_mars_bigtts":            {ID: "zh_male_ruyaqingnian_mars_bigtts", Name: "儒雅青年", Description: "儒雅的青年男声", Gender: "male"},
	"zh_male_baqiqingshu_mars_bigtts":             {ID: "zh_male_baqiqingshu_mars_bigtts", Name: "霸气青叔", Description: "霸气成熟的男声", Gender: "male"},
	"zh_male_qingcang_mars_bigtts":                {ID: "zh_male_qingcang_mars_bigtts", Name: "擎苍", Description: "威严霸气的男声", Gender: "male"},
	"zh_male_yangguangqingnian_mars_bigtts":       {ID: "zh_male_yangguangqingnian_mars_bigtts", Name: "活力小哥", Description: "活力四射的男声", Gender: "male"},
	"zh_male_fanjuanqingnian_mars_bigtts":         {ID: "zh_male_fanjuanqingnian_mars_bigtts", Name: "反卷青年", Description: "轻松随性的男声", Gender: "male"},
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

// LoadConfig 加载YAML配置文件
func LoadConfig() error {
	var err error
	configOnce.Do(func() {
		// 获取当前工作目录
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			err = fmt.Errorf("获取工作目录失败: %v", wdErr)
			return
		}

		// 构建配置文件路径
		fullPath := filepath.Join(wd, configFilePath)

		// 读取配置文件
		data, readErr := ioutil.ReadFile(fullPath)
		if readErr != nil {
			err = fmt.Errorf("读取配置文件失败 %s: %v", fullPath, readErr)
			return
		}

		// 解析YAML
		yamlConfig = &YAMLConfig{}
		if parseErr := yaml.Unmarshal(data, yamlConfig); parseErr != nil {
			err = fmt.Errorf("解析YAML配置失败: %v", parseErr)
			return
		}

		log.Printf("成功加载配置文件: %s", fullPath)
		log.Printf("主播身份码: %s", yamlConfig.IdCode)
		log.Printf("音色数量: %d", len(yamlConfig.Voices))
	})
	return err
}

// GetConfig 获取配置实例
func GetConfig() *YAMLConfig {
	if yamlConfig == nil {
		if err := LoadConfig(); err != nil {
			log.Printf("加载配置失败，使用默认配置: %v", err)
			return getDefaultConfig()
		}
	}
	return yamlConfig
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *YAMLConfig {
	return &YAMLConfig{
		IdCode: "",  // 默认为空，应从配置文件获取
		Voices: []Voice{
			{ID: DefaultVoice},
		},
	}
}

// GetIdCode 获取主播身份码
func GetIdCode() string {
	// 从YAML配置获取
	config := GetConfig()
	if config.IdCode != "" {
		return config.IdCode
	}
	
	// 最后返回空字符串，要求用户配置
	log.Println("警告: 未找到主播身份码配置，请在 config.yaml 中设置 id_code")
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

// GetVoices 获取所有音色配置
func GetVoices() []Voice {
	config := GetConfig()
	return config.Voices
}

// GetVoiceIDs 获取所有音色ID数组
func GetVoiceIDs() []string {
	voices := GetVoices()
	ids := make([]string, len(voices))
	for i, voice := range voices {
		ids[i] = voice.ID
	}
	return ids
}

// GetVoiceInfos 获取所有音色的完整信息
func GetVoiceInfos() []VoiceInfo {
	voices := GetVoices()
	infos := make([]VoiceInfo, len(voices))
	for i, voice := range voices {
		if info, exists := voiceNameMap[voice.ID]; exists {
			infos[i] = info
		} else {
			// 如果映射表中没有，创建一个基本的信息
			infos[i] = VoiceInfo{
				ID:          voice.ID,
				Name:        voice.ID, // 使用ID作为名称
				Description: "未知音色",
				Gender:      "unknown",
			}
		}
	}
	return infos
}

// GetRandomVoice 随机获取一个音色
func GetRandomVoice() Voice {
	voices := GetVoices()
	if len(voices) == 0 {
		// 返回默认音色
		return Voice{
			ID: DefaultVoice,
		}
	}
	index := rand.Intn(len(voices))
	return voices[index]
}

// GetRandomVoiceID 随机获取一个音色ID
func GetRandomVoiceID() string {
	voice := GetRandomVoice()
	return voice.ID
}

// GetRandomVoiceInfo 随机获取一个音色的完整信息
func GetRandomVoiceInfo() VoiceInfo {
	voice := GetRandomVoice()
	return GetVoiceInfoByID(voice.ID)
}

// GetVoiceByID 根据ID获取音色信息
func GetVoiceByID(id string) *Voice {
	voices := GetVoices()
	for _, voice := range voices {
		if voice.ID == id {
			return &voice
		}
	}
	return nil
}

// GetVoiceInfoByID 根据ID获取音色的完整信息
func GetVoiceInfoByID(id string) VoiceInfo {
	if info, exists := voiceNameMap[id]; exists {
		return info
	}
	// 如果映射表中没有，返回基本信息
	return VoiceInfo{
		ID:          id,
		Name:        id,
		Description: "未知音色",
		Gender:      "unknown",
	}
}

// GetVoiceInfosByGender 根据性别获取音色信息列表
func GetVoiceInfosByGender(gender string) []VoiceInfo {
	voices := GetVoices()
	var result []VoiceInfo
	for _, voice := range voices {
		info := GetVoiceInfoByID(voice.ID)
		if info.Gender == gender {
			result = append(result, info)
		}
	}
	return result
}

// GetVoicesByGender 根据性别获取音色列表（保持向后兼容）
func GetVoicesByGender(gender string) []Voice {
	infos := GetVoiceInfosByGender(gender)
	voices := make([]Voice, len(infos))
	for i, info := range infos {
		voices[i] = Voice{ID: info.ID}
	}
	return voices
}
