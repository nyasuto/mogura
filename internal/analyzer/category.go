package analyzer

type Category string

const (
	CategoryVideo    Category = "動画"
	CategoryImage    Category = "画像"
	CategoryCode     Category = "コード"
	CategoryDocument Category = "ドキュメント"
	CategoryCache    Category = "キャッシュ"
	CategoryArchive  Category = "アーカイブ"
	CategoryOther    Category = "その他"
)

var extToCategory = map[string]Category{
	// 動画
	".mp4":  CategoryVideo,
	".avi":  CategoryVideo,
	".mkv":  CategoryVideo,
	".mov":  CategoryVideo,
	".wmv":  CategoryVideo,
	".flv":  CategoryVideo,
	".webm": CategoryVideo,
	".m4v":  CategoryVideo,
	".mpg":  CategoryVideo,
	".mpeg": CategoryVideo,

	// 画像
	".jpg":  CategoryImage,
	".jpeg": CategoryImage,
	".png":  CategoryImage,
	".gif":  CategoryImage,
	".bmp":  CategoryImage,
	".svg":  CategoryImage,
	".webp": CategoryImage,
	".ico":  CategoryImage,
	".tiff": CategoryImage,
	".tif":  CategoryImage,
	".heic": CategoryImage,
	".raw":  CategoryImage,

	// コード
	".go":   CategoryCode,
	".py":   CategoryCode,
	".js":   CategoryCode,
	".ts":   CategoryCode,
	".jsx":  CategoryCode,
	".tsx":  CategoryCode,
	".java": CategoryCode,
	".c":    CategoryCode,
	".cpp":  CategoryCode,
	".h":    CategoryCode,
	".rs":   CategoryCode,
	".rb":   CategoryCode,
	".php":  CategoryCode,
	".cs":   CategoryCode,
	".swift": CategoryCode,
	".kt":  CategoryCode,
	".sh":  CategoryCode,
	".bash": CategoryCode,
	".zsh": CategoryCode,
	".sql": CategoryCode,
	".html": CategoryCode,
	".css":  CategoryCode,
	".scss": CategoryCode,
	".vue":  CategoryCode,

	// ドキュメント
	".pdf":  CategoryDocument,
	".doc":  CategoryDocument,
	".docx": CategoryDocument,
	".xls":  CategoryDocument,
	".xlsx": CategoryDocument,
	".ppt":  CategoryDocument,
	".pptx": CategoryDocument,
	".txt":  CategoryDocument,
	".md":   CategoryDocument,
	".csv":  CategoryDocument,
	".json": CategoryDocument,
	".xml":  CategoryDocument,
	".yaml": CategoryDocument,
	".yml":  CategoryDocument,
	".toml": CategoryDocument,

	// キャッシュ
	".tmp":   CategoryCache,
	".cache": CategoryCache,
	".log":   CategoryCache,
	".swp":   CategoryCache,
	".swo":   CategoryCache,
	".pyc":   CategoryCache,
	".o":     CategoryCache,
	".obj":   CategoryCache,
	".class": CategoryCache,
	".DS_Store": CategoryCache,

	// アーカイブ
	".zip":  CategoryArchive,
	".tar":  CategoryArchive,
	".gz":   CategoryArchive,
	".bz2":  CategoryArchive,
	".xz":   CategoryArchive,
	".7z":   CategoryArchive,
	".rar":  CategoryArchive,
	".dmg":  CategoryArchive,
	".iso":  CategoryArchive,
	".pkg":  CategoryArchive,
	".deb":  CategoryArchive,
	".rpm":  CategoryArchive,
}

func ClassifyExt(ext string) Category {
	if cat, ok := extToCategory[ext]; ok {
		return cat
	}
	return CategoryOther
}
