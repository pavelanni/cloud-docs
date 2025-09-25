local token = "?token=TOKEN_PLACEHOLDER"

function Link(link)
  -- Only modify local links (not http/https/mailto)
  if not link.target:match("^%a+://") and not link.target:match("^mailto:") then
    if not link.target:match("%?") then
      link.target = link.target .. token
    else
      link.target = link.target .. "&" .. token:sub(2)
    end
  else
    -- For external links, add target="_blank"
    if not link.target:match("^mailto:") then
      link.attributes = link.attributes or {}
      link.attributes["target"] = "_blank"
      -- Optionally, add rel="noopener noreferrer" for security
      link.attributes["rel"] = "noopener noreferrer"
    end
  end
  return link
end

function Image(img)
  -- Only modify local images
  if not img.src:match("^%a+://") then
    if not img.src:match("%?") then
      img.src = img.src .. token
    else
      img.src = img.src .. "&" .. token:sub(2)
    end
  end
  return img
end
