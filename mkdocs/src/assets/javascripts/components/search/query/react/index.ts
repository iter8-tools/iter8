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

import {
  MonoTypeOperatorFunction,
  Observable,
  combineLatest,
  fromEvent,
  merge,
  pipe
} from "rxjs"
import {
  delay,
  distinctUntilChanged,
  finalize,
  map,
  startWith,
  tap
} from "rxjs/operators"

import { watchElementFocus } from "browser"
import { SearchTransformFn, defaultTransform } from "integrations"

import { SearchQuery } from "../_"
import {
  resetSearchQueryPlaceholder,
  setSearchQueryPlaceholder
} from "../set"

/* ----------------------------------------------------------------------------
 * Helper types
 * ------------------------------------------------------------------------- */

/**
 * Watch options
 */
interface WatchOptions {
  transform?: SearchTransformFn        /* Transformation function */
}

/* ----------------------------------------------------------------------------
 * Functions
 * ------------------------------------------------------------------------- */

/**
 * Watch search query
 *
 * Note that the focus event which triggers re-reading the current query value
 * is delayed by `1ms` so the input's empty state is allowed to propagate.
 *
 * @param el - Search query element
 * @param options - Options
 *
 * @return Search query observable
 */
export function watchSearchQuery(
  el: HTMLInputElement, { transform }: WatchOptions = {}
): Observable<SearchQuery> {
  const fn = transform || defaultTransform

  /* Intercept keyboard events */
  const value$ = merge(
    fromEvent(el, "keyup"),
    fromEvent(el, "focus").pipe(delay(1))
  )
    .pipe(
      map(() => fn(el.value)),
      startWith(fn(el.value)),
      distinctUntilChanged()
    )

  /* Intercept focus events */
  const focus$ = watchElementFocus(el)

  /* Combine into single observable */
  return combineLatest([value$, focus$])
    .pipe(
      map(([value, focus]) => ({ value, focus }))
    )
}

/* ------------------------------------------------------------------------- */

/**
 * Apply search query
 *
 * @param el - Search query element
 *
 * @return Operator function
 */
export function applySearchQuery(
  el: HTMLInputElement
): MonoTypeOperatorFunction<SearchQuery> {
  return pipe(

    /* Hide placeholder when search is focused */
    tap(({ focus }) => {
      if (focus) {
        setSearchQueryPlaceholder(el, "")
      } else {
        resetSearchQueryPlaceholder(el)
      }
    }),

    /* Reset on complete or error */
    finalize(() => {
      resetSearchQueryPlaceholder(el)
    })
  )
}
