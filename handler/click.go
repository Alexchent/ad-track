package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Alexchent/ad-track/logic"
	"go.uber.org/zap"
)

func click(ctx context.Context, logicSvc *logic.Click, channel string, item map[string]interface{}) error {
	slog.InfoContext(ctx, "ClickData", zap.String("channel", channel), zap.Any("data", item))
	oaid, _ := item[logic.Oaid].(string)
	if oaid != "" {
		if err := logicSvc.SaveData(ctx, oaid, item); err != nil {
			return fmt.Errorf("save oaid or imei: %w", err)
		}
	}

	imei, _ := item[logic.Imei].(string)
	if imei != "" {
		if err := logicSvc.SaveData(ctx, imei, item); err != nil {
			return fmt.Errorf("save imei: %w", err)
		}
	}
	return nil
}
