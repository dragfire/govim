// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lsp

import (
	"context"

	"github.com/govim/govim/cmd/govim/internal/golang_org_x_tools/lsp/protocol"
	"github.com/govim/govim/cmd/govim/internal/golang_org_x_tools/span"
	"github.com/govim/govim/cmd/govim/internal/golang_org_x_tools/telemetry/log"
	"github.com/govim/govim/cmd/govim/internal/golang_org_x_tools/telemetry/tag"
)

func (s *Server) didChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	options := s.session.Options()
	if !options.WatchFileChanges {
		return nil
	}

	for _, change := range params.Changes {
		uri := span.NewURI(change.URI)

		switch change.Type {
		case protocol.Changed:
			view := s.session.ViewOf(uri)

			// If we have never seen this file before, there is nothing to do.
			if view.FindFile(ctx, uri) == nil {
				break
			}

			log.Print(ctx, "watched file changed", tag.Of("uri", uri))

			// If client has this file open, don't do anything. The client's contents
			// must remain the source of truth.
			if s.session.IsOpen(uri) {
				break
			}

			s.session.DidChangeOutOfBand(uri)

			// Refresh diagnostics to reflect updated file contents.
			s.Diagnostics(ctx, view, uri)
		case protocol.Created:
			log.Print(ctx, "watched file created", tag.Of("uri", uri))
		case protocol.Deleted:
			log.Print(ctx, "watched file deleted", tag.Of("uri", uri))
		}
	}

	return nil
}
