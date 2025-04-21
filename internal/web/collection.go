package web

import (
	"akira/internal/entity"
	"akira/internal/view/component/form"
	"akira/internal/view/page"
	"net/http"
	"strconv"
)

func (h *Handler) handleCreateCollectionRequest(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	session, err := h.session.GetSession(r.Context())
	if err != nil {
		err := entity.RequestError{}.Add("general", "unauthorized")
		return Render(w, r, page.SignIn(form.SignInProps{}, &err))
	}
	var volumesNum int
	if r.FormValue("total_volumes") != "" {
		var err error
		volumesNum, err = strconv.Atoi(r.FormValue("total_volumes"))
		if err != nil {
			var reqErr entity.RequestError
			reqErr.Add("total_volumes", "value invalid, must be a number")
			return Render(w, r, form.CreateCollection(form.CreateCollectionProps{
				Name:         r.FormValue("name"),
				TotalVolumes: 0,
				AutoSync:     r.FormValue("auto_sync") == "on",
				TrackPrices:  r.FormValue("track_price") == "on",
				TrackVolumes: r.FormValue("track_volumes") == "on",
				TrackReviews: r.FormValue("track_reviews") == "on",
			}, &reqErr))
		}
	}
	req := entity.CreateCollectionRequest{
		Name:         r.FormValue("name"),
		TotalVolumes: volumesNum,
		CrawlerOptions: entity.SyncOptions{
			AutoSync:        r.FormValue("auto_sync") == "on",
			TrackPrice:      r.FormValue("track_price") == "on",
			TrackNewVolumes: r.FormValue("track_volumes") == "on",
			TrackReviews:    r.FormValue("track_reviews") == "on",
		},
	}
	collection, err := h.collection.CreateCollection(session.UserID, req)
	if err != nil {
		if _, ok := err.(entity.RequestError); ok {
			err := err.(entity.RequestError)
			return Render(w, r, form.CreateCollection(form.CreateCollectionProps{
				Name:         req.Name,
				TotalVolumes: req.TotalVolumes,
				AutoSync:     req.CrawlerOptions.AutoSync,
				TrackPrices:  req.CrawlerOptions.TrackPrice,
				TrackVolumes: req.CrawlerOptions.TrackNewVolumes,
				TrackReviews: req.CrawlerOptions.TrackReviews,
			}, &err))
		}
		reqerr := entity.RequestError{}.Add("general", err.Error())
		h.logger.Error(r.Context(), "failed to create collection", err, map[string]any{
			"userID": session.UserID,
			"req":    req,
		})
		return Render(w, r, form.CreateCollection(form.CreateCollectionProps{
			Name:         req.Name,
			TotalVolumes: req.TotalVolumes,
			AutoSync:     req.CrawlerOptions.AutoSync,
			TrackPrices:  req.CrawlerOptions.TrackPrice,
			TrackVolumes: req.CrawlerOptions.TrackNewVolumes,
			TrackReviews: req.CrawlerOptions.TrackReviews,
		}, &reqerr))
	}
	return HxRedirect(w, r, "/collection/"+collection.Slug)
}
