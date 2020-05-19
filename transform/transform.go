package transform

import (
	"ai-dataset-tool/sql"
	"ai-dataset-tool/utils"
	"fmt"
)

func Test(annotationOutPath string, imageOutPath string) {
	//var wg sync.WaitGroup
	for _, lab := range sql.Labels {
		df := utils.TransformFile(lab.Image_path)
		fmt.Println(df.Name + "." + df.Suffix)
		//utils.DownloadIns.Goroutine_cnt <- 1
		//wg.Add(1)
		//go utils.DownloadIns.DGoroutine(&wg, df)
		utils.DownloadIns.GDownload(df)
		imageObj := NewImage(lab, 500, df)
		imageObj.initCellGroup(annotationOutPath, imageOutPath)
		imageObj.dispatchLabel()

		break
	}
	//wg.Wait()
}
