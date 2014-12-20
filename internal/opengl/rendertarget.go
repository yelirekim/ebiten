/*
Copyright 2014 Hajime Hoshi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package opengl

import (
	"errors"
	"fmt"
	"github.com/go-gl/gl"
	"github.com/hajimehoshi/ebiten/internal"
)

func orthoProjectionMatrix(left, right, bottom, top int) [4][4]float64 {
	e11 := float64(2) / float64(right-left)
	e22 := float64(2) / float64(top-bottom)
	e14 := -1 * float64(right+left) / float64(right-left)
	e24 := -1 * float64(top+bottom) / float64(top-bottom)

	return [4][4]float64{
		{e11, 0, 0, e14},
		{0, e22, 0, e24},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

type RenderTarget struct {
	framebuffer gl.Framebuffer
	width       int
	height      int
	flipY       bool
}

func NewZeroRenderTarget(width, height int) (*RenderTarget, error) {
	r := &RenderTarget{
		width:  width,
		height: height,
		flipY:  true,
	}
	return r, nil
}

func NewRenderTargetFromTexture(texture *Texture) (*RenderTarget, error) {
	framebuffer, err := createFramebuffer(texture.Native())
	if err != nil {
		return nil, err
	}
	return &RenderTarget{
		framebuffer: framebuffer,
		width:       texture.Width(),
		height:      texture.Height(),
	}, nil
}

func (r *RenderTarget) Width() int {
	return r.width
}

func (r *RenderTarget) Height() int {
	return r.height
}

func (r *RenderTarget) FlipY() bool {
	return r.flipY
}

func (r *RenderTarget) Dispose() {
	r.framebuffer.Delete()
}

func createFramebuffer(nativeTexture gl.Texture) (gl.Framebuffer, error) {
	// TODO: Does this affect the current rendering target?
	framebuffer := gl.GenFramebuffer()
	framebuffer.Bind()

	if err := framebufferTexture(nativeTexture); err != nil {
		return 0, err
	}

	return framebuffer, nil
}

func framebufferTexture(nativeTexture gl.Texture) error {
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, nativeTexture, 0)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		return errors.New("creating framebuffer failed")
	}

	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	return nil
}

func (r *RenderTarget) SetAsViewport() error {
	gl.Flush()
	r.framebuffer.Bind()
	err := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if err != gl.FRAMEBUFFER_COMPLETE {
		return errors.New(fmt.Sprintf("glBindFramebuffer failed: %d", err))
	}

	width := internal.AdjustSizeForTexture(r.width)
	height := internal.AdjustSizeForTexture(r.height)
	gl.Viewport(0, 0, width, height)
	return nil
}

func (r *RenderTarget) ProjectionMatrix() [4][4]float64 {
	width := internal.AdjustSizeForTexture(r.width)
	height := internal.AdjustSizeForTexture(r.height)
	m := orthoProjectionMatrix(0, width, 0, height)
	if r.flipY {
		m[1][1] *= -1
		m[1][3] += float64(r.height) / float64(internal.AdjustSizeForTexture(r.height)) * 2
	}
	return m
}
