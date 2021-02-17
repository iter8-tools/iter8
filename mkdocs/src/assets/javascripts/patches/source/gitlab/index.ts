/*
 * Copyright (c) 2016-2020 Martin Donath <martin.donath@squidfunk.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to
 * deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NON-INFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 */

import { ProjectSchema } from "gitlab"
import { Observable, from } from "rxjs"
import {
  defaultIfEmpty,
  filter,
  map,
  share,
  switchMap
} from "rxjs/operators"

import { round } from "utilities"

import { SourceFacts } from ".."

/* ----------------------------------------------------------------------------
 * Functions
 * ------------------------------------------------------------------------- */

/**
 * Fetch GitLab source facts
 *
 * @param base - GitLab base
 * @param project - GitLab project
 *
 * @return Source facts observable
 */
export function fetchSourceFactsFromGitLab(
  base: string, project: string
): Observable<SourceFacts> {
  const url = `https://${base}/api/v4/projects/${encodeURIComponent(project)}`
  return from(fetch(url))
    .pipe(
      filter(res => res.status === 200),
      switchMap(res => res.json()),
      map(({ star_count, forks_count }: ProjectSchema) => ([
        `${round(star_count)} Stars`,
        `${round(forks_count)} Forks`
      ])),
      defaultIfEmpty([]),
      share()
    )
}
