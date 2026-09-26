package core

import (
	"math/rand"
	"time"

	"github.com/mtgo-labs/mtgo/telegram"
	"github.com/mtgo-labs/mtgo/telegram/types"
)

// ProjectVersion — версия проекта Akari.
const ProjectVersion = "1.0.0"

// latinWords — список латинских слов для генерации имён устройств.
var latinWords = []string{
	"Absolutum", "Acies", "Aetas", "Aeternum", "Alea", "Akari", "Altus", "Amicus", "Animus",
	"Arcanus", "Ardent", "Astra", "Audax", "Aureus", "Bellum", "Bonus", "Caelum",
	"Candidus", "Cautus", "Celer", "Certo", "Clarus", "Cogito", "Comes", "Consto",
	"Corpus", "Cras", "Creo", "Cura", "Curro", "Custos", "Decus", "Deus", "Dignus",
	"Ducere", "Durus", "Eager", "Effulgeo", "Egregius", "Ensis", "Eques", "Equus",
	"Eximius", "Exitus", "Faber", "Fama", "Fatum", "Favus", "Ferox", "Fides",
	"Firmus", "Fluxus", "Fortis", "Fulgor", "Futurus", "Genius", "Gigno", "Gloria",
	"Gradus", "Gravis", "Habet", "Honor", "Humanus", "Ignis", "Imber", "Impetus",
	"Index", "Infinitus", "Innoxius", "Insignis", "Intellectus", "Invictus", "Ira",
	"Iustitia", "Labor", "Latus", "Laurus", "Legatus", "Lego", "Leo", "Levis",
	"Lex", "Liber", "Limen", "Luceo", "Lumen", "Lux", "Magnus", "Maior", "Maneo",
	"Mater", "Maximus", "Mens", "Meritus", "Miles", "Mirus", "Motus", "Mundus",
	"Natura", "Natus", "Nebula", "Nemo", "Niger", "Nitor", "Nobilis", "Nomen",
	"Novus", "Nox", "Nuntius", "Obsequens", "Occultus", "Omen", "Opus", "Orbis",
	"Ordo", "Ostendo", "Pax", "Pectus", "Pelle", "Pello", "Percipio", "Perfectus",
	"Perpetuus", "Persisto", "Piger", "Plerumque", "Plenus", "Polio", "Potens",
	"Praemium", "Primo", "Principium", "Probus", "Proelium", "Profecto", "Progenies",
	"Propter", "Prospicio", "Providus", "Prudentia", "Pugna", "Pulcher", "Purus",
	"Qualis", "Quasi", "Queror", "Quietus", "Quintus", "Quis", "Ratio", "Rectus",
	"Regnum", "Rem", "Repeto", "Requiro", "Res", "Rex", "Ritus", "Robur", "Rogatus",
	"Rosa", "Rudens", "Rursus", "Sacrum", "Salus", "Sanctus", "Sapientia", "Scio",
	"Scribo", "Securus", "Sedes", "Semper", "Senex", "Sentio", "Sequor", "Serenus",
	"Servo", "Sidera", "Signum", "Silens", "Simplex", "Sincerus", "Situs", "Socius",
	"Sol", "Solidus", "Solitudo", "Solvo", "Somnus", "Sonus", "Sors", "Speculum",
	"Spiritus", "Splendor", "Sponte", "Stabilis", "Status", "Stella", "Sternere",
	"Sto", "Strenuus", "Sub", "Subtilis", "Sufficio", "Sum", "Superbus", "Supremus",
	"Sursum", "Sustineo", "Talenta", "Tego", "Temeritas", "Tempus", "Tenax", "Teneo",
	"Terminus", "Terra", "Tertius", "Tolle", "Totus", "Tracto", "Tranquillus", "Tres",
	"Triumphus", "Trux", "Tueri", "Tumultus", "Ultio", "Ultra", "Unicus", "Unio",
	"Unus", "Urbs", "Usus", "Utilis", "Ut", "Vacuus", "Validus", "Vates", "Vectis",
	"Vehere", "Velox", "Ventus", "Verax", "Vereor", "Veritas", "Verto", "Vester",
	"Veteris", "Via", "Victor", "Vigil", "Vindex", "Vinctus", "Vinctus", "Vinum",
	"Vires", "Viridis", "Virtus", "Vis", "Vita", "Vito", "Vivus", "Voco", "Volens",
	"Voluptas", "Vox", "Vulgus",
	"Shadow", "Dream", "Kitsune", "Oni", "Yokai", "Kami", "Senpai", "Kohai",
	"Otaku", "Manga", "Anime", "Waifu", "Husbando", "Tsundere", "Yandere", "Kuudere",
	"DereDere", "Baka", "Sugoi", "Kawaii", "Sugoku", "Zettai", "Ryoken", "Shinigami",
	"Kamisama", "Youkai", "Obake", "Yurei", "Kappa", "Tengu", "Inari", "Tanuki",
	"Kitsunebi", "Onmyoji", "Shikigami", "Ofuda", "Omamori", "Torii", "Jinja",
	"Miko", "Kannushi", "Buddha", "Zen", "Satori", "Koan", "Haiku", "Tanka",
	"Bushido", "Samurai", "Ronin", "Ninja", "Shinobi", "Kunoichi", "Daimyo", "Shogun",
	"Katana", "Wakizashi", "Tanto", "Naginata", "Yumi", "Shuriken", "Kunai", "Fuma",
	"Chakra", "Jutsu", "Genjutsu", "Taijutsu", "Ninjutsu", "KekkeiGenkai", "Sharingan",
	"Byakugan", "Rinnegan", "Mangekyou", "Susanoo", "Amaterasu", "Tsukuyomi", "Kamui",
	"Hiraishin", "Rasengan", "Chidori", "Kirin", "Kaguya", "Hagoromo", "Indra", "Asura",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// LanguageConfig представляет пару языковых кодов.
type LanguageConfig struct {
	LangCode       string
	SystemLangCode string
}

// languages — реалистичные языковые пары для разных регионов.
var languages = []LanguageConfig{
	{LangCode: "en", SystemLangCode: "en-US"},
	{LangCode: "en", SystemLangCode: "en-GB"},
	{LangCode: "ru", SystemLangCode: "ru-RU"},
	{LangCode: "uk", SystemLangCode: "uk-UA"},
	{LangCode: "be", SystemLangCode: "be-BY"},
	{LangCode: "pl", SystemLangCode: "pl-PL"},
	{LangCode: "de", SystemLangCode: "de-DE"},
	{LangCode: "de", SystemLangCode: "de-AT"},
	{LangCode: "de", SystemLangCode: "de-CH"},
	{LangCode: "fr", SystemLangCode: "fr-FR"},
	{LangCode: "fr", SystemLangCode: "fr-CA"},
	{LangCode: "es", SystemLangCode: "es-ES"},
	{LangCode: "es", SystemLangCode: "es-MX"},
	{LangCode: "it", SystemLangCode: "it-IT"},
	{LangCode: "pt", SystemLangCode: "pt-PT"},
	{LangCode: "pt", SystemLangCode: "pt-BR"},
	{LangCode: "ja", SystemLangCode: "ja-JP"},
	{LangCode: "ko", SystemLangCode: "ko-KR"},
	{LangCode: "zh", SystemLangCode: "zh-CN"},
	{LangCode: "zh", SystemLangCode: "zh-TW"},
	{LangCode: "ar", SystemLangCode: "ar-SA"},
	{LangCode: "tr", SystemLangCode: "tr-TR"},
	{LangCode: "nl", SystemLangCode: "nl-NL"},
	{LangCode: "cs", SystemLangCode: "cs-CZ"},
	{LangCode: "sk", SystemLangCode: "sk-SK"},
	{LangCode: "hu", SystemLangCode: "hu-HU"},
	{LangCode: "ro", SystemLangCode: "ro-RO"},
	{LangCode: "bg", SystemLangCode: "bg-BG"},
	{LangCode: "hr", SystemLangCode: "hr-HR"},
	{LangCode: "sr", SystemLangCode: "sr-RS"},
	{LangCode: "sl", SystemLangCode: "sl-SI"},
	{LangCode: "et", SystemLangCode: "et-EE"},
	{LangCode: "lv", SystemLangCode: "lv-LV"},
	{LangCode: "lt", SystemLangCode: "lt-LT"},
	{LangCode: "fi", SystemLangCode: "fi-FI"},
	{LangCode: "sv", SystemLangCode: "sv-SE"},
	{LangCode: "da", SystemLangCode: "da-DK"},
	{LangCode: "no", SystemLangCode: "no-NO"},
	{LangCode: "el", SystemLangCode: "el-GR"},
	{LangCode: "he", SystemLangCode: "he-IL"},
	{LangCode: "hi", SystemLangCode: "hi-IN"},
	{LangCode: "th", SystemLangCode: "th-TH"},
	{LangCode: "vi", SystemLangCode: "vi-VN"},
	{LangCode: "id", SystemLangCode: "id-ID"},
	{LangCode: "ms", SystemLangCode: "ms-MY"},
}

// randomLanguage возвращает случайную языковую конфигурацию.
func randomLanguage() LanguageConfig {
	return languages[rand.Intn(len(languages))]
}

// GenerateAppName генерирует имя приложения из 3 случайных латинских слов.
// Пример: "Spiritus Gravis Fides"
func GenerateAppName() string {
	n := len(latinWords)
	if n == 0 {
		return "Akari Default"
	}
	indices := rand.Perm(n)[:3]
	return latinWords[indices[0]] + " " + latinWords[indices[1]] + " " + latinWords[indices[2]]
}

type DeviceProfile struct {
	DeviceModel    string
	SystemVersion  string
	LangPack       string
	ClientPlatform types.ClientPlatform
	PackageID      string
}

var windowsProfiles = []DeviceProfile{
	{DeviceModel: "Desktop", SystemVersion: "Windows 10", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Windows 11", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Windows 10 Pro", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Windows 11 Pro", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Windows 10 Enterprise", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
}

var linuxProfiles = []DeviceProfile{
	{DeviceModel: "Desktop", SystemVersion: "Ubuntu 22.04.4 LTS", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Ubuntu 24.04 LTS", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Fedora Linux 40", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Debian 12 (Bookworm)", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Arch Linux", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Linux Mint 21.3", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "openSUSE Tumbleweed", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Manjaro Linux", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Pop!_OS 22.04 LTS", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
	{DeviceModel: "Desktop", SystemVersion: "Elementary OS 7", LangPack: "tdesktop", ClientPlatform: types.ClientPlatformDesktop, PackageID: "org.telegram.desktop"},
}

var androidProfiles = []DeviceProfile{
	{DeviceModel: "Samsung Galaxy S24 Ultra", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Samsung Galaxy S23", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Samsung Galaxy S23 Ultra", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Google Pixel 8 Pro", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Google Pixel 8", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Google Pixel 7 Pro", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Xiaomi 14 Ultra", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "Xiaomi 13 Pro", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "OnePlus 12", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
	{DeviceModel: "OnePlus 11", SystemVersion: "Android 14", LangPack: "android", ClientPlatform: types.ClientPlatformAndroid, PackageID: "org.telegram.messenger"},
}

// GenerateRealisticDevice генерирует реалистичный профиль устройства.
// Вероятности: ~55% Windows, ~30% Linux, ~15% Android.
// Для десктопов с вероятностью 60% заменяет DeviceModel на GenerateAppName().
// AppVersion всегда равен версии проекта (ProjectVersion).
func GenerateRealisticDevice() telegram.DeviceConfig {
	r := rand.Float64()
	var profiles []DeviceProfile

	switch {
	case r < 0.55:
		profiles = windowsProfiles
	case r < 0.85:
		profiles = linuxProfiles
	default:
		profiles = androidProfiles
	}

	profile := profiles[rand.Intn(len(profiles))]

	if profile.ClientPlatform == types.ClientPlatformDesktop {
		if rand.Float64() < 0.6 {
			profile.DeviceModel = GenerateAppName()
		}
	}

	lang := randomLanguage()

	return telegram.DeviceConfig{
		DeviceModel:    profile.DeviceModel,
		SystemVersion:  profile.SystemVersion,
		AppVersion:     ProjectVersion,
		LangPack:       profile.LangPack,
		LangCode:       lang.LangCode,
		SystemLangCode: lang.SystemLangCode,
		ClientPlatform: profile.ClientPlatform,
		PackageID:      profile.PackageID,
		TZOffset:       10800,
	}
}

// GetOrCreateDeviceProfile возвращает профиль устройства из конфига или генерирует новый.
// Если в конфиге уже заданы device_model / system_version / app_version — использует их.
// Иначе генерирует новый профиль, сохраняет в конфиг и возвращает.
func GetOrCreateDeviceProfile(cfg *Config) telegram.DeviceConfig {
	cfg.mu.RLock()
	hasModel := cfg.DeviceModel != ""
	hasSysVer := cfg.SystemVersion != ""
	hasAppVer := cfg.AppVersion != ""
	cfg.mu.RUnlock()

	if hasModel && hasSysVer && hasAppVer {
		cfg.mu.RLock()
		defer cfg.mu.RUnlock()
		lang := randomLanguage()
		return telegram.DeviceConfig{
			DeviceModel:    cfg.DeviceModel,
			SystemVersion:  cfg.SystemVersion,
			AppVersion:     cfg.AppVersion,
			LangPack:       "tdesktop",
			LangCode:       lang.LangCode,
			SystemLangCode: lang.SystemLangCode,
			ClientPlatform: types.ClientPlatformDesktop,
			PackageID:      "org.telegram.desktop",
			TZOffset:       10800,
		}
	}

	device := GenerateRealisticDevice()

	cfg.mu.Lock()
	cfg.DeviceModel = device.DeviceModel
	cfg.SystemVersion = device.SystemVersion
	cfg.AppVersion = device.AppVersion
	cfg.mu.Unlock()

	if err := cfg.Save(); err != nil {
		_ = err
	}

	return device
}