import { Plus, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

type Props = {
  id: string;
  label: string;
  value: string[];
  placeholder: string;
  description?: string;
  addLabel: string;
  error?: string;
  onChange: (value: string[]) => void;
};

export function StringListField({
  id,
  label,
  value,
  placeholder,
  description,
  addLabel,
  error,
  onChange,
}: Props) {
  function update(index: number, nextValue: string) {
    onChange(value.map((item, itemIndex) => (itemIndex === index ? nextValue : item)));
  }

  return (
    <Field data-invalid={Boolean(error) || undefined}>
      <FieldLabel>{label}</FieldLabel>
      <div className="flex flex-col gap-2">
        {value.map((item, index) => (
          <div key={index} className="flex gap-2">
            <Input
              id={index === 0 ? id : undefined}
              className="font-mono"
              value={item}
              placeholder={placeholder}
              aria-invalid={Boolean(error)}
              onChange={(event) => update(index, event.target.value)}
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              disabled={value.length === 1}
              onClick={() => onChange(value.filter((_, itemIndex) => itemIndex !== index))}
            >
              <Trash2 />
            </Button>
          </div>
        ))}
      </div>
      <div className="flex items-center justify-between gap-3">
        {description ? <FieldDescription>{description}</FieldDescription> : <span />}
        <Button type="button" variant="outline" size="sm" onClick={() => onChange([...value, ""])}>
          <Plus data-icon="inline-start" />
          {addLabel}
        </Button>
      </div>
      <FieldError>{error}</FieldError>
    </Field>
  );
}
